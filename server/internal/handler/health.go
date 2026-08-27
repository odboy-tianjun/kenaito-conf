package handler

import (
	"kenaito-conf/server/internal/dao"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check 用于容器健康检查，数据库不可用时返回 503
func (h *HealthHandler) Check(c *gin.Context) {
	if dao.Engine == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "database not initialized"})
		return
	}
	if err := dao.Engine.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "database unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
