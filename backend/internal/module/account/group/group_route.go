package group

import (
	"micro-net-hub/internal/module/account/role"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitGroupRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	group := r.Group("/group")
	// 开启jwt认证中间件
	group.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	group.Use(role.CasbinMiddleware())
	{
		group.GET("/list", List)
		group.GET("/tree", GetTree)
		group.POST("/add", Add)
		group.POST("/update", Update)
		group.POST("/delete", Delete)
		group.POST("/adduser", AddUser)
		group.POST("/removeuser", RemoveUser)

		group.GET("/useringroup", UserInGroup)
		group.GET("/usernoingroup", UserNoInGroup)
	}

	return r
}
