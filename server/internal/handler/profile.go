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

type ProfileHandler struct{}

func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{}
}

func (h *ProfileHandler) List(c *gin.Context) {
	var profiles []model.Profile
	dao.Engine.Find(&profiles)
	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: profiles})
}

func (h *ProfileHandler) GetById(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var profile model.Profile
	has, _ := dao.Engine.ID(id).Get(&profile)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "环境不存在"})
		return
	}

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: profile})
}

type AddProfileRequest struct {
	Name      string `json:"name" binding:"required"`
	Alias     string `json:"alias"`
	IsDefault bool   `json:"isDefault"`
}

func (h *ProfileHandler) Add(c *gin.Context) {
	var req AddProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if !userInfo.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "超级管理员才能操作"})
		return
	}

	var existing model.Profile
	if has, _ := dao.Engine.Where("name = ?", req.Name).Get(&existing); has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 1002, Msg: "环境已存在"})
		return
	}

	profile := model.Profile{
		Name:      req.Name,
		Alias:     req.Alias,
		IsDefault: req.IsDefault,
		Enabled:   true,
		CreatedAt: func() *time.Time { t := time.Now(); return &t }(),
	}

	dao.Engine.Insert(&profile)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

type UpdateProfileRequest struct {
	Id        int    `json:"id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Alias     string `json:"alias"`
	IsDefault bool   `json:"isDefault"`
}

func (h *ProfileHandler) Update(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if !userInfo.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "超级管理员才能操作"})
		return
	}

	var profile model.Profile
	has, _ := dao.Engine.ID(req.Id).Get(&profile)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "环境不存在"})
		return
	}

	if profile.Name != req.Name {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "环境标识不可修改"})
		return
	}

	profile.Alias = req.Alias
	now := time.Now()
	profile.UpdatedAt = &now
	dao.Engine.ID(profile.Id).Cols("alias", "is_default", "updated_at").Update(&profile)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

func (h *ProfileHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userInfo := middleware.GetUserInfo(c)
	if !userInfo.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "超级管理员才能操作"})
		return
	}

	var profile model.Profile
	has, _ := dao.Engine.ID(id).Get(&profile)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "环境不存在"})
		return
	}

	if profile.IsDefault {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 1003, Msg: "默认环境不能删除"})
		return
	}

	dao.Engine.ID(id).Delete(&model.Profile{})

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0})
}

func (h *ProfileHandler) Switch(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userInfo := middleware.GetUserInfo(c)
	if !userInfo.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "超级管理员才能操作"})
		return
	}

	var profile model.Profile
	has, _ := dao.Engine.ID(id).Get(&profile)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "环境不存在"})
		return
	}

	profile.Enabled = !profile.Enabled
	dao.Engine.ID(profile.Id).Cols("enabled").Update(&profile)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

func (h *ProfileHandler) Options(c *gin.Context) {
	var profiles []model.Profile
	dao.Engine.Where("enabled = 1").Find(&profiles)

	options := make([]model.SelectOption, 0)
	for _, p := range profiles {
		options = append(options, model.SelectOption{Value: p.Name, Label: p.Alias + "(" + p.Name + ")"})
	}

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: options})
}
