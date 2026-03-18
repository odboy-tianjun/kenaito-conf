package handler

import (
	"crypto/rand"
	"encoding/hex"
	"kenaito-conf/server/internal/dao"
	"kenaito-conf/server/internal/middleware"
	"kenaito-conf/server/internal/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AppHandler struct{}

func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

type AppListRequest struct {
	PageNo   int    `form:"pageNo" binding:"required"`
	PageSize int    `form:"pageSize" binding:"required"`
	AppType  string `form:"appType"`
}

func (h *AppHandler) List(c *gin.Context) {
	var req AppListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	userInfo := middleware.GetUserInfo(c)
	query := dao.Engine.Where("enabled = 1")

	if !userInfo.IsSuperAdmin() {
		if userInfo.IsGroupAdmin() && userInfo.GroupId != nil {
			query = query.And("group_id = ?", *userInfo.GroupId)
		} else {
			userId, _ := strconv.Atoi(userInfo.Id)
			query = query.And("(owner_id = ? OR id IN (SELECT app_id FROM user_permissions WHERE user_id = ?))", userId, userId)
		}
	}

	var apps []model.Application
	whereClause := "enabled = 1"
	if !userInfo.IsSuperAdmin() {
		if userInfo.IsGroupAdmin() && userInfo.GroupId != nil {
			whereClause += " AND group_id = " + strconv.Itoa(*userInfo.GroupId)
		} else {
			userId, _ := strconv.Atoi(userInfo.Id)
			whereClause += " AND (owner_id = " + strconv.Itoa(userId) + " OR id IN (SELECT app_id FROM user_permissions WHERE user_id = " + strconv.Itoa(userId) + "))"
		}
	}
	if req.AppType != "" {
		whereClause += " AND service_id = '" + req.AppType + "'"
	}
	
	total, _ := dao.Engine.Where(whereClause).Count(&model.Application{})
	dao.Engine.Where(whereClause).Limit(req.PageSize, (req.PageNo-1)*req.PageSize).Find(&apps)

	// load group names - batch query to avoid N+1
	groupIds := make(map[int]bool)
	for _, app := range apps {
		if app.GroupId != nil {
			groupIds[*app.GroupId] = true
		}
	}
	groupMap := make(map[int]string)
	if len(groupIds) > 0 {
		ids := make([]int, 0, len(groupIds))
		for gid := range groupIds {
			ids = append(ids, gid)
		}
		var groups []model.BusinessGroup
		dao.Engine.In("id", ids).Find(&groups)
		for _, g := range groups {
			groupMap[g.Id] = g.Name
		}
	}
	for i := range apps {
		if apps[i].GroupId != nil {
			apps[i].GroupName = groupMap[*apps[i].GroupId]
		}
	}

	c.JSON(http.StatusOK, model.PageResult{PageNo: req.PageNo, PageSize: req.PageSize, Total: total, Data: apps})
}

func (h *AppHandler) GetById(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var app model.Application
	has, _ := dao.Engine.ID(id).Get(&app)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "应用不存在"})
		return
	}
	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: app})
}

type AddAppRequest struct {
	Id        int    `json:"id"`
	GroupId   *int   `json:"groupId"`
	ParentId  *int   `json:"parentId"`
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code" binding:"required"`
	ServiceId string `json:"serviceId"`
	OwnerId   *int   `json:"ownerId"`
	Remarks   string `json:"remarks"`
}

func (h *AppHandler) Add(c *gin.Context) {
	var req AddAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	userInfo := middleware.GetUserInfo(c)

	// check if code exists
	var existing model.Application
	if has, _ := dao.Engine.Where("code = ?", req.Code).Get(&existing); has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 1002, Msg: "应用已存在"})
		return
	}

	ownerId := req.OwnerId
	if ownerId == nil {
		uid, _ := strconv.Atoi(userInfo.Id)
		ownerId = &uid
	}

	var owner model.User
	if has, _ := dao.Engine.ID(*ownerId).Get(&owner); !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "负责人不存在"})
		return
	}

	app := model.Application{
		GroupId:   req.GroupId,
		ParentId:  req.ParentId,
		Name:      req.Name,
		Code:      req.Code,
		ServiceId: req.ServiceId,
		OwnerId:   ownerId,
		OwnerName: owner.Name,
		Remarks:   req.Remarks,
		Enabled:   true,
		CreatedAt: func() *time.Time { t := time.Now(); return &t }(),
		CreatedBy: userInfo.Name,
	}

	if req.ParentId != nil {
		var parent model.Application
		if has, _ := dao.Engine.ID(*req.ParentId).Get(&parent); has {
			app.GroupId = parent.GroupId
			app.IsModule = true
			app.Code = parent.Code + "-" + req.Code
		}
	}

	dao.Engine.Insert(&app)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

func (h *AppHandler) Update(c *gin.Context) {
	var req AddAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	if req.Id == 0 {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "id参数缺失"})
		return
	}

	var app model.Application
	has, _ := dao.Engine.ID(req.Id).Get(&app)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "应用不存在"})
		return
	}

	userInfo := middleware.GetUserInfo(c)

	app.Name = req.Name
	app.ServiceId = req.ServiceId
	app.Remarks = req.Remarks
	if req.OwnerId != nil {
		if app.OwnerId == nil || *req.OwnerId != *app.OwnerId {
			var owner model.User
			if has, _ := dao.Engine.ID(*req.OwnerId).Get(&owner); has {
				app.OwnerName = owner.Name
			}
		}
		app.OwnerId = req.OwnerId
	}
	app.UpdatedBy = userInfo.Name
	now := time.Now()
	app.UpdatedAt = &now

	dao.Engine.ID(app.Id).Cols("name", "service_id", "remarks", "owner_id", "owner_name", "updated_by", "updated_at").Update(&app)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

func (h *AppHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userInfo := middleware.GetUserInfo(c)

	if !userInfo.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "超级管理员才能删除"})
		return
	}

	var app model.Application
	has, _ := dao.Engine.ID(id).Get(&app)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "应用不存在"})
		return
	}

	app.Enabled = false
	app.UpdatedBy = userInfo.Name
	now := time.Now()
	app.UpdatedAt = &now
	dao.Engine.ID(app.Id).Cols("enabled", "updated_by", "updated_at").Update(&app)

	// soft delete related configs
	dao.Engine.Where("app_id = ?", id).Cols("enabled").Update(&model.AppConfig{Enabled: false})

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0})
}

func (h *AppHandler) Switch(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userInfo := middleware.GetUserInfo(c)

	var app model.Application
	has, _ := dao.Engine.ID(id).Get(&app)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "应用不存在"})
		return
	}

	app.Enabled = !app.Enabled
	app.UpdatedBy = userInfo.Name
	now := time.Now()
	app.UpdatedAt = &now
	dao.Engine.ID(app.Id).Cols("enabled", "updated_by", "updated_at").Update(&app)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

func (h *AppHandler) Options(c *gin.Context) {
	userInfo := middleware.GetUserInfo(c)
	query := dao.Engine.Where("enabled = 1")

	if !userInfo.IsSuperAdmin() {
		if userInfo.IsGroupAdmin() && userInfo.GroupId != nil {
			query = query.And("group_id = ?", *userInfo.GroupId)
		} else {
			userId, _ := strconv.Atoi(userInfo.Id)
			query = query.And("owner_id = ?", userId)
		}
	}

	var apps []model.Application
	query.Find(&apps)

	options := make([]model.SelectOption, 0)
	for _, app := range apps {
		options = append(options, model.SelectOption{Value: strconv.Itoa(app.Id), Label: app.Name})
	}

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: options})
}

func (h *AppHandler) GenerateToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userInfo := middleware.GetUserInfo(c)

	var app model.Application
	has, _ := dao.Engine.ID(id).Get(&app)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "应用不存在"})
		return
	}

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 500, Msg: "生成 Token 失败"})
		return
	}
	token := hex.EncodeToString(bytes)

	app.Token = token
	app.UpdatedBy = userInfo.Name
	now := time.Now()
	app.UpdatedAt = &now
	dao.Engine.ID(app.Id).Cols("token", "updated_by", "updated_at").Update(&app)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: token})
}

func (h *AppHandler) DeleteToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userInfo := middleware.GetUserInfo(c)

	var app model.Application
	has, _ := dao.Engine.ID(id).Get(&app)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "应用不存在"})
		return
	}

	app.Token = ""
	app.UpdatedBy = userInfo.Name
	now := time.Now()
	app.UpdatedAt = &now
	dao.Engine.ID(app.Id).Cols("token", "updated_by", "updated_at").Update(&app)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}
