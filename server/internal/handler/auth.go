package handler

import (
	"kenaito-conf/server/internal/middleware"
	"kenaito-conf/server/internal/model"
	"kenaito-conf/server/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	token, userInfo, err := h.authService.Login(req.UserName, req.Password)
	if err != nil {
		c.JSON(http.StatusOK, model.WrapperResponse{Code: 400, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.WrapperResponse{
		Code: 0,
		Data: gin.H{"token": token, "userInfo": userInfo},
	})
}

func (h *AuthHandler) GetLoginUserInfo(c *gin.Context) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, model.WrapperResponse{Code: 401, Msg: "未登录"})
		return
	}
	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: userInfo})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0})
}
