package handler

import (
	"kenaito-conf/server/internal/middleware"
	"kenaito-conf/server/internal/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IndexHandler struct{}

func NewIndexHandler() *IndexHandler {
	return &IndexHandler{}
}

func (h *IndexHandler) GetMenus(c *gin.Context) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, model.WrapperResponse{Code: 401, Msg: "未登录"})
		return
	}

	menus := make([]model.Menu, 0)

	// 环境管理
	if userInfo.IsSuperAdmin() {
		menuEnv := model.NewMenu("环境管理", "fa-cubes", true)
		menuEnv.AddSubMenu("环境管理", "&#xe61d;", "/profile/list")
		menus = append(menus, menuEnv)
	}

	// 业务组管理
	if userInfo.IsSuperAdmin() {
		menuGroup := model.NewMenu("业务组管理", "fa-cubes", true)
		menuGroup.AddSubMenu("业务组管理", "&#xe613;", "/group/list")
		menus = append(menus, menuGroup)
	}

	// 用户管理
	if userInfo.IsSuperAdmin() || userInfo.IsGroupAdmin() {
		menuUser := model.NewMenu("用户管理", "fa-cubes", true)
		menuUser.AddSubMenu("用户管理", "&#xe612;", "/user/list")
		menus = append(menus, menuUser)
	}

	// 应用管理
	menuApp := model.NewMenu("应用管理", "fa-cubes", true)
	menuApp.AddSubMenu("应用管理", "&#xe63b;", "/app/list")
	menus = append(menus, menuApp)

	// 配置管理
	menuConfig := model.NewMenu("配置管理", "fa-cubes", true)
	menuConfig.AddSubMenu("配置管理", "&#xe60a;", "/app/config/list")
	menus = append(menus, menuConfig)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: menus})
}
