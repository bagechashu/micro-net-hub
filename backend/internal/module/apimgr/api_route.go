package apimgr

import (
	"micro-net-hub/internal/module/account/role"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitApiRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	api := r.Group("/api")
	// 开启jwt认证中间件
	api.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	api.Use(role.CasbinMiddleware())
	{
		api.GET("/tree", GetTree)
		api.GET("/list", List)
		api.POST("/add", Add)
		api.POST("/update", Update)
		api.POST("/delete", Delete)
	}

	return r
}
