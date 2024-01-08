package routes

import (
	"micro-net-hub/internal/server/handler"
	"micro-net-hub/internal/server/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitApiRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	api := r.Group("/api")
	// 开启jwt认证中间件
	api.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	api.Use(middleware.CasbinMiddleware())
	{
		var h handler.ApiHandler
		api.GET("/tree", h.GetTree)
		api.GET("/list", h.List)
		api.POST("/add", h.Add)
		api.POST("/update", h.Update)
		api.POST("/delete", h.Delete)
	}

	return r
}
