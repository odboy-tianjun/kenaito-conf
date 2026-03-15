package handler

import (
	"kenaito-conf/server/internal/dao"
	"kenaito-conf/server/internal/middleware"
	"kenaito-conf/server/internal/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct{}

func NewGroupHandler() *GroupHandler {
	return &GroupHandler{}
}

func (h *GroupHandler) List(c *gin.Context) {
	var groups []model.BusinessGroup
	dao.Engine.Find(&groups)
	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: groups})
}

func (h *GroupHandler) GetById(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var group model.BusinessGroup
	has, _ := dao.Engine.ID(id).Get(&group)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "业务组不存在"})
		return
	}
	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: group})
}

type AddGroupRequest struct {
	Id        int    `json:"id"`
	Name      string `json:"name" binding:"required"`
	MasterUid *int   `json:"masterUid" binding:"required"`
	Remarks   string `json:"remarks"`
}

func (h *GroupHandler) Add(c *gin.Context) {
	var req AddGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if !userInfo.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "超级管理员才能操作"})
		return
	}

	var existing model.BusinessGroup
	if has, _ := dao.Engine.Where("name = ?", req.Name).Get(&existing); has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 1002, Msg: "组名已存在"})
		return
	}

	var master model.User
	if has, _ := dao.Engine.ID(*req.MasterUid).Get(&master); !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "管理员不存在"})
		return
	}

	group := model.BusinessGroup{
		Name:      req.Name,
		Master:    master.Name,
		MasterUid: req.MasterUid,
		Remarks:   req.Remarks,
		Enabled:   true,
		CreatedAt: func() *time.Time { t := time.Now(); return &t }(),
		CreatedBy: userInfo.Name,
	}

	dao.Engine.Insert(&group)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0})
}

func (h *GroupHandler) Update(c *gin.Context) {
	var req AddGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if !userInfo.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "超级管理员才能操作"})
		return
	}

	if req.Id == 0 {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "id参数缺失"})
		return
	}

	var group model.BusinessGroup
	has, _ := dao.Engine.ID(req.Id).Get(&group)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "业务组不存在"})
		return
	}

	group.Name = req.Name
	group.Remarks = req.Remarks
	if req.MasterUid != nil && (group.MasterUid == nil || *req.MasterUid != *group.MasterUid) {
		var master model.User
		if has, _ := dao.Engine.ID(*req.MasterUid).Get(&master); has {
			group.Master = master.Name
			group.MasterUid = req.MasterUid
		}
	}
	group.UpdatedBy = userInfo.Name
	now := time.Now()
	group.UpdatedAt = &now

	dao.Engine.ID(group.Id).Cols("name", "remarks", "master", "master_uid", "updated_by", "updated_at").Update(&group)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0})
}

func (h *GroupHandler) Switch(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userInfo := middleware.GetUserInfo(c)
	if !userInfo.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "超级管理员才能操作"})
		return
	}

	var group model.BusinessGroup
	has, _ := dao.Engine.ID(id).Get(&group)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "业务组不存在"})
		return
	}

	group.Enabled = !group.Enabled
	now := time.Now()
	group.UpdatedAt = &now
	group.UpdatedBy = userInfo.Name
	dao.Engine.ID(group.Id).Cols("enabled", "updated_at", "updated_by").Update(&group)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

func (h *GroupHandler) Options(c *gin.Context) {
	var groups []model.BusinessGroup
	dao.Engine.Where("enabled = 1").Find(&groups)

	options := make([]model.SelectOption, 0)
	for _, g := range groups {
		options = append(options, model.SelectOption{Value: strconv.Itoa(g.Id), Label: g.Name})
	}

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: options})
}
