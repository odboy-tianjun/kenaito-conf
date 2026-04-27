package handler

import (
	"kenaito-conf/server/internal/dao"
	"kenaito-conf/server/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct{}

func NewClientHandler() *ClientHandler {
	return &ClientHandler{}
}

func (h *ClientHandler) ListByConfigId(c *gin.Context) {
	configId, _ := strconv.Atoi(c.Query("configId"))
	if configId == 0 {
		c.JSON(http.StatusBadRequest, model.WrapperResponse{Code: 400, Msg: "参数错误"})
		return
	}

	var instances []model.ClientInstance
	dao.Engine.Where("app_config_id = ?", configId).OrderBy("last_fetch_at DESC").Find(&instances)

	c.JSON(http.StatusOK, model.WrapperResponse{Code: 0, Data: instances})
}
