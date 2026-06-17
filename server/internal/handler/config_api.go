package handler

import (
	"fmt"
	"kenaito-conf/server/internal/dao"
	"kenaito-conf/server/internal/logger"
	"kenaito-conf/server/internal/model"
	"kenaito-conf/server/internal/service"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// getClientIP 获取客户端 IPv4 地址，将 IPv6 环回地址转为 127.0.0.1
func getClientIP(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "" {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	if parsed.IsLoopback() {
		return "127.0.0.1"
	}
	if v4 := parsed.To4(); v4 != nil {
		return v4.String()
	}
	return ip
}

type ConfigApiHandler struct {
	parseService *service.ConfigParseService
}

func NewConfigApiHandler(parse *service.ConfigParseService) *ConfigApiHandler {
	return &ConfigApiHandler{parseService: parse}
}

type FetchAllRequest struct {
	AppName      string `json:"appName" binding:"required"`
	Env          string `json:"env" binding:"required"`
	Version      string `json:"version"`
	IgnoreGlobal bool   `json:"ignoreGlobal"`
	Token        string `json:"token"`
}

func (h *ConfigApiHandler) FetchAllConfigs(c *gin.Context) {
	var req FetchAllRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	version := req.Version
	if version == "" {
		version = "latest"
	}

	if req.AppName == "" || req.Env == "" {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数缺失"})
		return
	}

	// find app by code
	var app model.Application
	has, _ := dao.Engine.Where("code = ? AND enabled = 1", req.AppName).Get(&app)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 1001, Msg: "配置不存在"})
		return
	}

	// check if app's group is enabled
	if app.GroupId != nil {
		var group model.BusinessGroup
		has, _ := dao.Engine.ID(*app.GroupId).Get(&group)
		if !has || !group.Enabled {
			c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "业务组已禁用"})
			return
		}
	}

	// validate token if app has token configured
	if app.Token != "" && app.Token != req.Token {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "Token 无效"})
		return
	}

	result := make(map[string]interface{})
	var configIds []int

	// load global configs
	if !req.IgnoreGlobal && app.GroupId != nil {
		var globalConfigs []model.AppConfig
		dao.Engine.Where("env = ? AND version = ? AND enabled = 1 AND group_id = ? AND (app_id is null OR app_id = 0)",
			req.Env, version, *app.GroupId).Find(&globalConfigs)
		for _, cfg := range globalConfigs {
			configIds = append(configIds, cfg.Id)
			kvMap := h.parseService.ParseConfigToKVMap(&cfg)
			for k, v := range kvMap {
				result[k] = v
			}
		}
	}

	// load app configs
	var appConfigs []model.AppConfig
	dao.Engine.Where("env = ? AND version = ? AND enabled = 1 AND app_id = ?",
		req.Env, version, app.Id).Find(&appConfigs)
	for _, cfg := range appConfigs {
		configIds = append(configIds, cfg.Id)
		kvMap := h.parseService.ParseConfigToKVMap(&cfg)
		for k, v := range kvMap {
			result[k] = v
		}
	}

	if len(result) == 0 {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 1001, Msg: "配置不存在"})
		return
	}

	result["latest.fetch.timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)

	// record client instances asynchronously
	go h.recordClientInstances(getClientIP(c), configIds, req.AppName, req.Env, "fetchAll")

	c.JSON(http.StatusOK, result)
}

type FetchChangedRequest struct {
	AppName      string `json:"appName" binding:"required"`
	Env          string `json:"env" binding:"required"`
	Version      string `json:"version"`
	LastTime     int64  `json:"lastTime" binding:"required"`
	IgnoreGlobal bool   `json:"ignoreGlobal"`
	FetchAll     bool   `json:"fetchAll"`
	Token        string `json:"token"`
}

func (h *ConfigApiHandler) FetchChangedConfigs(c *gin.Context) {
	var req FetchChangedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	var app model.Application
	has, _ := dao.Engine.Where("code = ? AND enabled = 1", req.AppName).Get(&app)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 1001, Msg: "应用不存在"})
		return
	}

	// validate token if app has token configured
	if app.Token != "" && app.Token != req.Token {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "Token 无效"})
		return
	}

	version := req.Version
	if version == "" {
		version = "latest"
	}

	lastUpdateTime := time.UnixMilli(req.LastTime)

	result := make(map[string]string)
	var configIds []int

	// 查询已删除的配置（enabled=0 且 updated_at > lastTime），返回 deletedKeys
	var deletedKeys []string
	if !req.IgnoreGlobal && app.GroupId != nil {
		var deletedGlobalConfigs []model.AppConfig
		dao.Engine.Where("env = ? AND version = ? AND enabled = 0 AND group_id = ? AND (app_id is null OR app_id = 0) AND updated_at > ?",
			req.Env, version, *app.GroupId, lastUpdateTime).Find(&deletedGlobalConfigs)
		for _, cfg := range deletedGlobalConfigs {
			kvMap := h.parseService.ParseConfigToKVMap(&cfg)
			for k := range kvMap {
				deletedKeys = append(deletedKeys, k)
			}
		}
	}
	var deletedAppConfigs []model.AppConfig
	dao.Engine.Where("env = ? AND version = ? AND enabled = 0 AND app_id = ? AND updated_at > ?",
		req.Env, version, app.Id, lastUpdateTime).Find(&deletedAppConfigs)
	for _, cfg := range deletedAppConfigs {
		kvMap := h.parseService.ParseConfigToKVMap(&cfg)
		for k := range kvMap {
			deletedKeys = append(deletedKeys, k)
		}
	}
	if len(deletedKeys) > 0 {
		result["deletedKeys"] = strings.Join(deletedKeys, ",")
	}

	// load global changed configs
	if !req.IgnoreGlobal && app.GroupId != nil {
		var globalConfigs []model.AppConfig
		dao.Engine.Where("env = ? AND version = ? AND enabled = 1 AND group_id = ? AND (app_id is null OR app_id = 0) AND updated_at > ?",
			req.Env, version, *app.GroupId, lastUpdateTime).Find(&globalConfigs)
		for _, cfg := range globalConfigs {
			configIds = append(configIds, cfg.Id)
			h.addChangedConfigs(result, &cfg, req.FetchAll)
		}
	}

	// load app changed configs
	var appConfigs1 []model.AppConfig
	dao.Engine.Where("env = ? AND version = ? AND enabled = 1 AND app_id = ? AND updated_at > ?",
		req.Env, version, app.Id, lastUpdateTime).Find(&appConfigs1)
	for _, cfg := range appConfigs1 {
		configIds = append(configIds, cfg.Id)
		h.addChangedConfigs(result, &cfg, req.FetchAll)
	}

	result["latest.fetch.timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)

	// query all config ids for this app to record client instances
	var allConfigIds []int
	if !req.IgnoreGlobal && app.GroupId != nil {
		var globalConfigs []model.AppConfig
		dao.Engine.Where("env = ? AND version = ? AND enabled = 1 AND group_id = ? AND (app_id is null OR app_id = 0)",
			req.Env, version, *app.GroupId).Find(&globalConfigs)
		for _, cfg := range globalConfigs {
			allConfigIds = append(allConfigIds, cfg.Id)
		}
	}
	var appConfigs2 []model.AppConfig
	dao.Engine.Where("env = ? AND version = ? AND enabled = 1 AND app_id = ?",
		req.Env, version, app.Id).Find(&appConfigs2)
	for _, cfg := range appConfigs2 {
		allConfigIds = append(allConfigIds, cfg.Id)
	}

	// record client instances asynchronously
	// only update last_fetch_at if there are actual changes
	fetchType := ""
	if len(configIds) > 0 {
		fetchType = "fetchChanged"
	}
	go h.recordClientInstances(getClientIP(c), allConfigIds, req.AppName, req.Env, fetchType)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: result})
}

func (h *ConfigApiHandler) addChangedConfigs(result map[string]string, config *model.AppConfig, fetchAll bool) {
	kvMap := h.parseService.ParseConfigToKVMap(config)
	for k, v := range kvMap {
		result[k] = fmt.Sprintf("%v", v)
	}
}

func (h *ConfigApiHandler) recordClientInstances(ip string, configIds []int, appName, env, fetchType string) {
	if len(configIds) == 0 || ip == "" {
		return
	}
	now := time.Now()

	// 批量查询已存在的记录
	var existingList []model.ClientInstance
	dao.Engine.Where("app_name = ? AND env = ? AND ip = ?", appName, env, ip).
		In("app_config_id", configIds).
		Find(&existingList)

	existingMap := make(map[int]*model.ClientInstance, len(existingList))
	for i := range existingList {
		existingMap[existingList[i].AppConfigId] = &existingList[i]
	}

	var toInsert []model.ClientInstance
	for _, configId := range configIds {
		if existing, ok := existingMap[configId]; ok {
			existing.LastPingAt = &now
			existing.UpdatedAt = &now
			if fetchType != "" {
				existing.LastFetchAt = &now
				existing.LastFetchType = fetchType
				dao.Engine.ID(existing.Id).Cols("last_ping_at", "last_fetch_at", "last_fetch_type", "updated_at").Update(existing)
			} else {
				dao.Engine.ID(existing.Id).Cols("last_ping_at", "updated_at").Update(existing)
			}
		} else {
			instance := model.ClientInstance{
				AppConfigId: configId,
				AppName:     appName,
				Env:         env,
				Ip:          ip,
				LastPingAt:  &now,
				CreatedAt:   &now,
				UpdatedAt:   &now,
			}
			if fetchType != "" {
				instance.LastFetchAt = &now
				instance.LastFetchType = fetchType
			}
			toInsert = append(toInsert, instance)
		}
	}

	if len(toInsert) > 0 {
		if _, err := dao.Engine.Insert(&toInsert); err != nil {
			logger.SugaredLogger.Errorf("[recordClientInstances] batch insert error: %v", err)
		}
	}
}

// func (h *ConfigApiHandler) Ping(c *gin.Context) {
// 	c.String(http.StatusOK, "200")
// }
