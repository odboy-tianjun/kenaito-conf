package handler

import (
	"kenaito-conf/server/internal/dao"
	"kenaito-conf/server/internal/middleware"
	"kenaito-conf/server/internal/model"
	"kenaito-conf/server/internal/service"
	"kenaito-conf/server/internal/util"
	"crypto/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	permService *service.PermissionService
}

func NewUserHandler(perm *service.PermissionService) *UserHandler {
	return &UserHandler{permService: perm}
}

func (h *UserHandler) List(c *gin.Context) {
	groupIdStr := c.Query("groupId")
	userInfo := middleware.GetUserInfo(c)

	var users []model.User
	query := dao.Engine.Where("1=1")

	if !userInfo.IsSuperAdmin() && userInfo.GroupId != nil {
		query = query.And("group_id = ?", *userInfo.GroupId)
	}
	if groupIdStr != "" {
		groupId, _ := strconv.Atoi(groupIdStr)
		query = query.And("group_id = ?", groupId)
	}

	query.Find(&users)

	// load group names
	var groups []model.BusinessGroup
	dao.Engine.Find(&groups)
	groupMap := make(map[int]string)
	for _, g := range groups {
		groupMap[g.Id] = g.Name
	}

	for i := range users {
		if users[i].GroupId != nil {
			users[i].GroupName = groupMap[*users[i].GroupId]
		}
	}

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: users})
}

type AddUserRequest struct {
	Name      string `json:"name" binding:"required"`
	Mobile    string `json:"mobile" binding:"required"`
	Email     string `json:"email"`
	Type      string `json:"type" binding:"required"`
	GroupId   *int   `json:"groupId"`
	GroupName string `json:"groupName"`
}

func (h *UserHandler) Add(c *gin.Context) {
	var req AddUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	userInfo := middleware.GetUserInfo(c)

	// validate type - only superAdmin can create superAdmin
	if req.Type != "user" && req.Type != "groupAdmin" && req.Type != "superAdmin" {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "无效的用户类型"})
		return
	}
	if req.Type == "superAdmin" && !userInfo.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "只有超级管理员才能创建超级管理员"})
		return
	}

	// check name exists
	var existing model.User
	if has, _ := dao.Engine.Where("name = ?", req.Name).Get(&existing); has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 1001, Msg: "用户名已存在"})
		return
	}
	if has, _ := dao.Engine.Where("mobile = ?", req.Mobile).Get(&existing); has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 1001, Msg: "手机号已存在"})
		return
	}

	groupId := req.GroupId
	if groupId == nil && req.GroupName != "" {
		var group model.BusinessGroup
		if has, _ := dao.Engine.Where("name = ?", req.GroupName).Get(&group); has {
			groupId = &group.Id
		} else {
			group = model.BusinessGroup{
				Name:      req.GroupName,
				Enabled:   true,
				CreatedAt: func() *time.Time { t := time.Now(); return &t }(),
				CreatedBy: userInfo.Name,
			}
			dao.Engine.Insert(&group)
			groupId = &group.Id
		}
	}

	password := "admin123"
	hashedPassword, _ := util.HashPassword(password)

	user := model.User{
		Name:      req.Name,
		Mobile:    req.Mobile,
		Email:     req.Email,
		Type:      req.Type,
		GroupId:   groupId,
		Password:  hashedPassword,
		Enabled:   true,
		CreatedAt: func() *time.Time { t := time.Now(); return &t }(),
	}

	dao.Engine.Insert(&user)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: user})
}

type UpdateUserRequest struct {
	Email       string `json:"email"`
	Mobile      string `json:"mobile"`
	Password    string `json:"password"`
	OldPassword string `json:"oldPassword"`
	GroupId     *int   `json:"groupId"`
}

func (h *UserHandler) Update(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	userInfo := middleware.GetUserInfo(c)
	userId, _ := strconv.Atoi(userInfo.Id)

	var user model.User
	has, _ := dao.Engine.ID(userId).Get(&user)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "用户不存在"})
		return
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Mobile != "" {
		user.Mobile = req.Mobile
	}
	if req.Password != "" {
		if err := util.VerifyPassword(user.Password, req.OldPassword); err != nil {
			c.JSON(http.StatusOK, model.WrapperResponse{Code: 1001, Msg: "原密码不正确"})
			return
		}
		hashedPassword, _ := util.HashPassword(req.Password)
		user.Password = hashedPassword
	}
	if req.GroupId != nil && *req.GroupId > 0 {
		user.GroupId = req.GroupId
	}
	user.UpdatedAt = func() *time.Time { t := time.Now(); return &t }()

	dao.Engine.ID(user.Id).Cols("email", "mobile", "password", "group_id", "updated_at").Update(&user)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

type AdminUpdateUserRequest struct {
	Id        int    `json:"id" binding:"required"`
	Email     string `json:"email"`
	Mobile    string `json:"mobile"`
	Type      string `json:"type"`
	GroupId   *int   `json:"groupId"`
	GroupName string `json:"groupName"`
}

func (h *UserHandler) AdminUpdate(c *gin.Context) {
	var req AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if !userInfo.IsSuperAdmin() {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: "只有超级管理员才能编辑用户"})
		return
	}

	var user model.User
	has, _ := dao.Engine.ID(req.Id).Get(&user)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "用户不存在"})
		return
	}

	// superAdmin 不能修改类型和所属组
	if user.Type == "superAdmin" {
		if req.Type != "" && req.Type != "superAdmin" {
			c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "不能修改超级管理员的类型"})
			return
		}
		if req.GroupId != nil || req.GroupName != "" {
			c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "不能修改超级管理员的所属组"})
			return
		}
	}

	if req.Mobile != "" {
		user.Mobile = req.Mobile
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Type != "" && user.Type != "superAdmin" {
		user.Type = req.Type
	}

	// 处理所属组
	if user.Type != "superAdmin" {
		if req.GroupId != nil && *req.GroupId > 0 {
			user.GroupId = req.GroupId
		} else if req.GroupName != "" {
			var group model.BusinessGroup
			has, _ := dao.Engine.Where("name = ?", req.GroupName).Get(&group)
			if !has {
				newGroup := model.BusinessGroup{
					Name:      req.GroupName,
					Enabled:   true,
					CreatedAt: func() *time.Time { t := time.Now(); return &t }(),
					CreatedBy: userInfo.Name,
				}
				dao.Engine.Insert(&newGroup)
				user.GroupId = &newGroup.Id
			} else {
				user.GroupId = &group.Id
			}
		}
	}

	user.UpdatedAt = func() *time.Time { t := time.Now(); return &t }()
	dao.Engine.ID(user.Id).Cols("mobile", "email", "type", "group_id", "updated_at").Update(&user)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userInfo := middleware.GetUserInfo(c)

	var user model.User
	has, _ := dao.Engine.ID(id).Get(&user)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "用户不存在"})
		return
	}

	if err := h.permService.HasPermissionForUser(userInfo, &user); err != nil {
		c.JSON(http.StatusForbidden, model.WrapperResponse{Code: 403, Msg: err.Error()})
		return
	}

	dao.Engine.ID(id).Delete(&model.User{})

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var user model.User
	has, _ := dao.Engine.ID(id).Get(&user)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "用户不存在"})
		return
	}

	newPassword := randomString(8)
	hashedPassword, _ := util.HashPassword(newPassword)
	user.Password = hashedPassword
	now := time.Now()
	user.UpdatedAt = &now
	dao.Engine.ID(user.Id).Cols("password", "updated_at").Update(&user)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: newPassword})
}

func (h *UserHandler) Switch(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var user model.User
	has, _ := dao.Engine.ID(id).Get(&user)
	if !has {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 404, Msg: "用户不存在"})
		return
	}

	// 用户名为 admin 的用户不能被禁用
	if user.Name == "admin" && user.Enabled {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: "admin 用户不能被禁用"})
		return
	}

	user.Enabled = !user.Enabled
	dao.Engine.ID(user.Id).Cols("enabled").Update(&user)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: true})
}

func (h *UserHandler) Options(c *gin.Context) {
	groupIdStr := c.Query("groupId")
	var users []model.User
	query := dao.Engine.Where("enabled = 1")
	if groupIdStr != "" {
		groupId, _ := strconv.Atoi(groupIdStr)
		query = query.And("group_id = ?", groupId)
	}
	query.Find(&users)

	options := make([]model.SelectOption, 0)
	for _, u := range users {
		options = append(options, model.SelectOption{Value: strconv.Itoa(u.Id), Label: u.Name})
	}

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: options})
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}
	return string(b)
}
