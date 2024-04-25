package role

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitRoleRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	role := r.Group("/role")
	// 开启jwt认证中间件
	role.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	role.Use(CasbinMiddleware())
	{
		role.GET("/list", RoleList)
		role.POST("/add", RoleAdd)
		role.POST("/update", RoleUpdate)
		role.POST("/delete", RoleDelete)

		role.GET("/getmenulist", RoleGetMenuList)
		role.GET("/getapilist", RoleGetApiList)
		role.POST("/updatemenus", RoleUpdateMenus)
		role.POST("/updateapis", RoleUpdateApis)
	}
	return r
}
