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
		api.GET("/tree", handler.Api.GetTree)
		api.GET("/list", handler.Api.List)
		api.POST("/add", handler.Api.Add)
		api.POST("/update", handler.Api.Update)
		api.POST("/delete", handler.Api.Delete)
	}

	return r
}
