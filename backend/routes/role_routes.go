package routes

import (
	"github.com/eryajf/go-ldap-admin/controller"
	"github.com/eryajf/go-ldap-admin/middleware"

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
		role.GET("/list", controller.Role.List)
		role.POST("/add", controller.Role.Add)
		role.POST("/update", controller.Role.Update)
		role.POST("/delete", controller.Role.Delete)

		role.GET("/getmenulist", controller.Role.GetMenuList)
		role.GET("/getapilist", controller.Role.GetApiList)
		role.POST("/updatemenus", controller.Role.UpdateMenus)
		role.POST("/updateapis", controller.Role.UpdateApis)
	}
	return r
}
