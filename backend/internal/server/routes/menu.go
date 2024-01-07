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
		menu.GET("/tree", handler.Menu.GetTree)
		menu.GET("/access/tree", handler.Menu.GetAccessTree)
		menu.POST("/add", handler.Menu.Add)
		menu.POST("/update", handler.Menu.Update)
		menu.POST("/delete", handler.Menu.Delete)
	}

	return r
}
