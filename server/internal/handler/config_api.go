package handler

import (
	"fmt"
	"kenaito-conf/server/internal/dao"
	"kenaito-conf/server/internal/model"
	"kenaito-conf/server/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ConfigApiHandler struct {
	parseService  *service.ConfigParseService
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

	// load global configs
	if !req.IgnoreGlobal && app.GroupId != nil {
		var globalConfigs []model.AppConfig
		dao.Engine.Where("env = ? AND version = ? AND enabled = 1 AND group_id = ? AND (app_id is null OR app_id = 0)",
			req.Env, version, *app.GroupId).Find(&globalConfigs)
		for _, cfg := range globalConfigs {
			kvMap := h.parseService.ParseConfigToKVMap(&cfg)
			for k, v := range kvMap {
				result[k] = v
			}
		}
	}

	// load parent app configs
	if app.ParentId != nil {
		var parentConfigs []model.AppConfig
		dao.Engine.Where("env = ? AND version = ? AND enabled = 1 AND app_id = ?",
			req.Env, version, *app.ParentId).Find(&parentConfigs)
		for _, cfg := range parentConfigs {
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

	// load global changed configs
	if !req.IgnoreGlobal && app.GroupId != nil {
		var globalConfigs []model.AppConfig
		dao.Engine.Where("env = ? AND version = ? AND enabled = 1 AND group_id = ? AND (app_id is null OR app_id = 0) AND updated_at > ?",
			req.Env, version, *app.GroupId, lastUpdateTime).Find(&globalConfigs)
		for _, cfg := range globalConfigs {
			h.addChangedConfigs(result, &cfg, req.FetchAll)
		}
	}

	// load parent app changed configs
	if app.ParentId != nil {
		var parentConfigs []model.AppConfig
		dao.Engine.Where("env = ? AND version = ? AND enabled = 1 AND app_id = ? AND updated_at > ?",
			req.Env, version, *app.ParentId, lastUpdateTime).Find(&parentConfigs)
		for _, cfg := range parentConfigs {
			h.addChangedConfigs(result, &cfg, req.FetchAll)
		}
	}

	// load app changed configs
	var appConfigs []model.AppConfig
	dao.Engine.Where("env = ? AND version = ? AND enabled = 1 AND app_id = ? AND updated_at > ?",
		req.Env, version, app.Id, lastUpdateTime).Find(&appConfigs)
	for _, cfg := range appConfigs {
		h.addChangedConfigs(result, &cfg, req.FetchAll)
	}

	result["latest.fetch.timestamp"] = strconv.FormatInt(time.Now().UnixMilli(), 10)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: result})
}

func (h *ConfigApiHandler) addChangedConfigs(result map[string]string, config *model.AppConfig, fetchAll bool) {
	kvMap := h.parseService.ParseConfigToKVMap(config)
	for k, v := range kvMap {
		result[k] = fmt.Sprintf("%v", v)
	}
}

func (h *ConfigApiHandler) Ping(c *gin.Context) {
	c.String(http.StatusOK, "200")
}
