package menu

import (
	"micro-net-hub/internal/module/account/role"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitMenuRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	menu := r.Group("/menu")
	// 开启jwt认证中间件
	menu.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	menu.Use(role.CasbinMiddleware())
	{
		menu.GET("/tree", GetTree)
		menu.GET("/access/tree", GetAccessTree)
		menu.POST("/add", Add)
		menu.POST("/update", Update)
		menu.POST("/delete", Delete)
	}

	return r
}
