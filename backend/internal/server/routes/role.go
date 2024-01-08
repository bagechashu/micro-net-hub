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
		var h handler.RoleHandler
		role.GET("/list", h.List)
		role.POST("/add", h.Add)
		role.POST("/update", h.Update)
		role.POST("/delete", h.Delete)

		role.GET("/getmenulist", h.GetMenuList)
		role.GET("/getapilist", h.GetApiList)
		role.POST("/updatemenus", h.UpdateMenus)
		role.POST("/updateapis", h.UpdateApis)
	}
	return r
}
