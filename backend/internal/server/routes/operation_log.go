package routes

import (
	"micro-net-hub/internal/server/handler"
	"micro-net-hub/internal/server/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitOperationLogRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	operation_log := r.Group("/log")
	// 开启jwt认证中间件
	operation_log.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	operation_log.Use(middleware.CasbinMiddleware())
	{
		var h handler.OperationLogHandler
		operation_log.GET("/operation/list", h.List)
		operation_log.POST("/operation/delete", h.Delete)
	}
	return r
}
