package handler

import (
	"kenaito-conf/server/internal/dao"
	"kenaito-conf/server/internal/middleware"
	"kenaito-conf/server/internal/model"
	"kenaito-conf/server/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	cryptoService *service.CryptoService
	parseService  *service.ConfigParseService
	permService   *service.PermissionService
}

func NewConfigHandler(crypto *service.CryptoService, parse *service.ConfigParseService, perm *service.PermissionService) *ConfigHandler {
	return &ConfigHandler{cryptoService: crypto, parseService: parse, permService: perm}
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
	query := dao.Engine.Where("enabled = 1")

	if req.Env != "" {
		query = query.And("env = ?", req.Env)
	}
	if req.AppId != nil {
		if *req.AppId <= 0 {
			query = query.And("(app_id is null OR app_id = 0)")
		} else {
			var app model.Application
			if has, _ := dao.Engine.ID(*req.AppId).Get(&app); has {
				query = query.And("app_code like ?", app.Code+"%")
			}
		}
	}

	// permission filter
	if !userInfo.IsSuperAdmin() {
		if userInfo.IsGroupAdmin() && userInfo.GroupId != nil {
			query = query.And("group_id = ?", *userInfo.GroupId)
		} else if len(userInfo.GrantAppIds) > 0 {
			query = query.In("app_id", userInfo.GrantAppIds)
		} else {
			c.JSON(http.StatusOK, model.PageResult{PageNo: req.PageNo, PageSize: req.PageSize, Total: 0, Data: []interface{}{}})
			return
		}
	}

	var configs []model.AppConfig
	total, _ := query.Count(&model.AppConfig{})
	query.Limit(req.PageSize, (req.PageNo-1)*req.PageSize).Find(&configs)

	// build app names
	for i := range configs {
		configs[i].AppNames = h.buildAppNames(&configs[i])
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

	// encrypt sensitive values
	h.encryptIfRequired(&config)

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

	h.encryptIfRequired(&config)

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

func (h *ConfigHandler) encryptIfRequired(config *model.AppConfig) {
	// simplified: check if contents contain sensitive keys
	// In production, parse the config and encrypt values with {Cipher} prefix
	// For now, skip encryption logic
}
