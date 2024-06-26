package sitenav

import (
	"micro-net-hub/internal/module/account/role"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitSiteNavRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	sitenav := r.Group("/sitenav")
	{
		sitenav.GET("/getnav", GetNav)
	}

	group := sitenav.Group("/group")
	group.Use(authMiddleware.MiddlewareFunc())
	group.Use(role.CasbinMiddleware())
	{
		group.POST("/add", AddNavGroup)
		group.POST("/update", UpdateNavGroup)
		group.POST("/delete", DeleteNavGroup)
	}

	site := sitenav.Group("/site")
	site.Use(authMiddleware.MiddlewareFunc())
	site.Use(role.CasbinMiddleware())
	{
		site.POST("/add", AddNavSite)
		site.POST("/update", UpdateNavSite)
		site.POST("/delete", DeleteNavSite)
	}
	return r
}
