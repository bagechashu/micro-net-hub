package routes

import (
	"micro-net-hub/internal/server/handler"
	"micro-net-hub/internal/server/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitRoleRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	role := r.Group("/role")
	// 开启jwt认证中间件
	role.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	role.Use(middleware.CasbinMiddleware())
	{
		role.GET("/list", handler.Role.List)
		role.POST("/add", handler.Role.Add)
		role.POST("/update", handler.Role.Update)
		role.POST("/delete", handler.Role.Delete)

		role.GET("/getmenulist", handler.Role.GetMenuList)
		role.GET("/getapilist", handler.Role.GetApiList)
		role.POST("/updatemenus", handler.Role.UpdateMenus)
		role.POST("/updateapis", handler.Role.UpdateApis)
	}
	return r
}
