package handler

import (
	"kenaito-conf/server/internal/dao"
	"kenaito-conf/server/internal/middleware"
	"kenaito-conf/server/internal/model"
	"kenaito-conf/server/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	parseService *service.ConfigParseService
	permService  *service.PermissionService
}

func NewConfigHandler(parse *service.ConfigParseService, perm *service.PermissionService) *ConfigHandler {
	return &ConfigHandler{parseService: parse, permService: perm}
}

type ConfigListRequest struct {
	PageNo   int    `form:"pageNo" binding:"required"`
	PageSize int    `form:"pageSize" binding:"required"`
	Env      string `form:"env"`
	AppId    *int   `form:"appId"`
}

func (h *ConfigHandler) List(c *gin.Context) {
	var req ConfigListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	userInfo := middleware.GetUserInfo(c)

	// 检查权限
	if !userInfo.IsSuperAdmin() {
		if !userInfo.IsGroupAdmin() && len(userInfo.GrantAppIds) == 0 {
			c.JSON(http.StatusOK, model.PageResult{PageNo: req.PageNo, PageSize: req.PageSize, Total: 0, Data: []interface{}{}})
			return
		}
	}

	// 构建查询条件
	conditions := []string{"enabled = 1"}
	args := make([]interface{}, 0, 4)
	if req.Env != "" {
		conditions = append(conditions, "env = ?")
		args = append(args, req.Env)
	}
	if req.AppId != nil {
		if *req.AppId <= 0 {
			conditions = append(conditions, "(app_id is null OR app_id = 0)")
		} else {
			conditions = append(conditions, "app_id = ?")
			args = append(args, *req.AppId)
		}
	}
	if !userInfo.IsSuperAdmin() {
		if userInfo.IsGroupAdmin() && userInfo.GroupId != nil {
			conditions = append(conditions, "group_id = ?")
			args = append(args, *userInfo.GroupId)
		} else if len(userInfo.GrantAppIds) > 0 {
			ids := make([]string, len(userInfo.GrantAppIds))
			for i, id := range userInfo.GrantAppIds {
				ids[i] = strconv.Itoa(id)
			}
			conditions = append(conditions, "app_id in ("+strings.Join(ids, ",")+")")
		}
	}

	whereClause := strings.Join(conditions, " AND ")

	var configs []model.AppConfig
	total, _ := dao.Engine.Where(whereClause, args...).Count(&model.AppConfig{})
	dao.Engine.Where(whereClause, args...).Limit(req.PageSize, (req.PageNo-1)*req.PageSize).Find(&configs)

	// build app names (batch query) and collect configIds for client counts
	configIds := make([]int, 0, len(configs))
	appIds := make(map[int]bool)
	for i := range configs {
		if configs[i].AppId != nil && *configs[i].AppId > 0 {
			appIds[*configs[i].AppId] = true
		}
		configIds = append(configIds, configs[i].Id)
	}
	if len(appIds) > 0 {
		idList := make([]int, 0, len(appIds))
		for id := range appIds {
			idList = append(idList, id)
		}
		var apps []model.Application
		dao.Engine.In("id", idList).Find(&apps)
		appMap := make(map[int]string, len(apps))
		for _, app := range apps {
			appMap[app.Id] = app.Name + "(" + app.Code + ")"
		}
		for i := range configs {
			if configs[i].Global || configs[i].AppId == nil || *configs[i].AppId == 0 {
				configs[i].AppNames = "全局配置"
			} else {
				configs[i].AppNames = appMap[*configs[i].AppId]
			}
		}
	} else {
		for i := range configs {
			if configs[i].Global || configs[i].AppId == nil || *configs[i].AppId == 0 {
				configs[i].AppNames = "全局配置"
			}
		}
	}
	// batch query client counts
	if len(configIds) > 0 {
		type countResult struct {
			AppConfigId int `xorm:"app_config_id"`
			Cnt         int `xorm:"cnt"`
		}
		var results []countResult
		dao.Engine.Table(&model.ClientInstance{}).
			Select("app_config_id, COUNT(*) as cnt").
			In("app_config_id", configIds).
			GroupBy("app_config_id").
			Find(&results)
		countMap := make(map[int]int, len(results))
		for _, r := range results {
			countMap[r.AppConfigId] = r.Cnt
		}
		for i := range configs {
			configs[i].ClientCount = countMap[configs[i].Id]
		}
	}

	c.JSON(http.StatusOK, model.PageResult{PageNo: req.PageNo, PageSize: req.PageSize, Total: total, Data: configs})
}

func (h *ConfigHandler) GetById(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var config model.AppConfig
	has, _ := dao.Engine.ID(id).Get(&config)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "配置不存在"})
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if err := h.permService.RequirePermission(userInfo, config.GroupId, config.AppId, "RO"); err != nil {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: err.Error()})
		return
	}

	config.AppNames = h.buildAppNames(&config)
	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: config})
}

type AddConfigRequest struct {
	Id       int    `json:"id"`
	GroupId  *int   `json:"groupId"`
	AppId    *int   `json:"appId"`
	Env      string `json:"env" binding:"required"`
	Version  string `json:"version"`
	Name     string `json:"name"`
	Type     int    `json:"type" binding:"required"`
	Contents string `json:"contents"`
	Global   bool   `json:"global"`
}

func (h *ConfigHandler) Add(c *gin.Context) {
	var req AddConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if req.Global {
		if req.GroupId == nil {
			c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "全局配置需要指定所属业务组"})
			return
		}
		if err := h.permService.RequirePermission(userInfo, req.GroupId, nil, "RW"); err != nil {
			c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: err.Error()})
			return
		}
	}

	if !req.Global && req.AppId == nil {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "非全局绑定应用不能为空"})
		return
	}

	var app model.Application
	if req.AppId != nil {
		if has, _ := dao.Engine.ID(*req.AppId).Get(&app); !has {
			c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "应用不存在"})
			return
		}
	}

	config := model.AppConfig{
		GroupId:   req.GroupId,
		AppId:     req.AppId,
		Env:       req.Env,
		Version:   req.Version,
		Name:      req.Name,
		Type:      req.Type,
		Contents:  req.Contents,
		Enabled:   true,
		CreatedAt: func() *time.Time { t := time.Now(); return &t }(),
		CreatedBy: userInfo.Name,
		UpdatedAt: func() *time.Time { t := time.Now(); return &t }(),
		UpdatedBy: userInfo.Name,
	}

	if req.AppId != nil {
		config.AppCode = app.Code
		if !req.Global {
			config.GroupId = app.GroupId
		}
	}

	if _, err := dao.Engine.Insert(&config); err != nil {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 500, Msg: "保存失败"})
		return
	}

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

func (h *ConfigHandler) Update(c *gin.Context) {
	var req AddConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	if req.Id == 0 {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "id参数缺失"})
		return
	}

	var config model.AppConfig
	has, _ := dao.Engine.ID(req.Id).Get(&config)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "配置不存在"})
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if err := h.permService.RequirePermission(userInfo, config.GroupId, config.AppId, "RW"); err != nil {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: err.Error()})
		return
	}

	// save history
	h.saveHistory(&config, userInfo.Name)

	originalContents := config.Contents
	config.Contents = req.Contents
	config.Version = req.Version
	config.UpdatedBy = userInfo.Name
	now := time.Now()
	config.UpdatedAt = &now

	dao.Engine.ID(config.Id).Cols("contents", "version", "updated_by", "updated_at").Update(&config)

	// TODO: publish config change event (ZK removed, could use webhook or other mechanism)
	_ = originalContents

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

func (h *ConfigHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var config model.AppConfig
	has, _ := dao.Engine.ID(id).Get(&config)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "配置不存在"})
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if err := h.permService.RequirePermission(userInfo, config.GroupId, config.AppId, "RW"); err != nil {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: err.Error()})
		return
	}

	config.Enabled = false
	config.UpdatedBy = userInfo.Name
	now := time.Now()
	config.UpdatedAt = &now
	dao.Engine.ID(config.Id).Cols("enabled", "updated_by", "updated_at").Update(&config)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0})
}

func (h *ConfigHandler) GetHistory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var histories []model.AppConfigHistory
	dao.Engine.Where("origin_id = ?", id).OrderBy("id desc").Limit(5).Find(&histories)

	if len(histories) > 0 {
		var currentConfig model.AppConfig
		if has, _ := dao.Engine.ID(histories[0].OriginId).Get(&currentConfig); has {
			for i := range histories {
				histories[i].ActiveContents = currentConfig.Contents
			}
		}
	}

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: histories})
}

func (h *ConfigHandler) Rollback(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var history model.AppConfigHistory
	has, _ := dao.Engine.ID(id).Get(&history)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "历史记录不存在"})
		return
	}

	var config model.AppConfig
	if has, _ := dao.Engine.ID(history.OriginId).Get(&config); has {
		userInfo := middleware.GetUserInfo(c)
		if err := h.permService.RequirePermission(userInfo, config.GroupId, config.AppId, "RW"); err != nil {
			c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: err.Error()})
			return
		}

		config.Contents = history.Contents
		config.UpdatedBy = userInfo.Name
		now := time.Now()
		config.UpdatedAt = &now
		dao.Engine.ID(config.Id).Cols("contents", "updated_by", "updated_at").Update(&config)
	}

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0})
}

func (h *ConfigHandler) buildAppNames(config *model.AppConfig) string {
	if config.Global || (config.AppId == nil || *config.AppId == 0) {
		return "全局配置"
	}
	var app model.Application
	if has, _ := dao.Engine.ID(*config.AppId).Get(&app); has {
		return app.Name + "(" + app.Code + ")"
	}
	return ""
}

func (h *ConfigHandler) saveHistory(config *model.AppConfig, operator string) {
	history := model.AppConfigHistory{
		OriginId:  config.Id,
		AppId:     config.AppId,
		AppNames:  h.buildAppNames(config),
		Env:       config.Env,
		Version:   config.Version,
		Name:      config.Name,
		Type:      config.Type,
		Contents:  config.Contents,
		CreatedAt: func() *time.Time { t := time.Now(); return &t }(),
		CreatedBy: operator,
	}
	dao.Engine.Insert(&history)
}
