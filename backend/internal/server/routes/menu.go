package routes

import (
	"micro-net-hub/internal/server/handler"
	"micro-net-hub/internal/server/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitMenuRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	menu := r.Group("/menu")
	// 开启jwt认证中间件
	menu.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	menu.Use(middleware.CasbinMiddleware())
	{
		var h handler.MenuHandler
		menu.GET("/tree", h.GetTree)
		menu.GET("/access/tree", h.GetAccessTree)
		menu.POST("/add", h.Add)
		menu.POST("/update", h.Update)
		menu.POST("/delete", h.Delete)
	}

	return r
}
