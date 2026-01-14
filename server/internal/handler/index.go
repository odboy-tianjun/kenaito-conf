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

	// 全局管理菜单
	menu1 := model.NewMenu("全局管理", "fa-cubes", false)
	if userInfo.IsSuperAdmin() {
		menu1.AddSubMenu("环境管理", "&#xe61d;", "/profile/list")
	}
	menu1.AddSubMenu("业务组管理", "&#xe613;", "/group/list")
	if userInfo.IsSuperAdmin() || userInfo.IsGroupAdmin() {
		menu1.AddSubMenu("用户管理", "&#xe612;", "/user/list")
	}
	menu1.AddSubMenu("应用管理", "&#xe63b;", "/app/list")
	menus = append(menus, menu1)

	// 配置管理菜单
	menu2 := model.NewMenu("配置管理", "fa-cubes", true)
	menu2.AddSubMenu("配置列表", "&#xe60a;", "/config/list")
	menu2.AddSubMenu("新建配置", "&#xe63c;", "/config/add")
	menus = append(menus, menu2)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: menus})
}
