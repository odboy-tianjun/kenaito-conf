package middleware

import (
	"kenaito-conf/server/internal/model"
	"kenaito-conf/server/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JWTAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		userInfo, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "登录已过期"})
			c.Abort()
			return
		}

		c.Set("userInfo", userInfo)
		c.Next()
	}
}

func GetUserInfo(c *gin.Context) *model.LoginUserInfo {
	val, exists := c.Get("userInfo")
	if !exists {
		return nil
	}
	return val.(*model.LoginUserInfo)
}
