package routes

import (
	"micro-net-hub/internal/server/handler"
	"micro-net-hub/internal/server/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitSiteNavRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	var h handler.SiteNavHandler

	sitenav := r.Group("/sitenav")
	{
		sitenav.GET("/getnav", h.GetNav)
	}

	group := sitenav.Group("/group")
	group.Use(authMiddleware.MiddlewareFunc())
	group.Use(middleware.CasbinMiddleware())
	{
		group.POST("/add", h.AddNavGroup)
		group.POST("/update", h.UpdateNavGroup)
		group.POST("/delete", h.DeleteNavGroup)
	}

	site := sitenav.Group("/site")
	site.Use(authMiddleware.MiddlewareFunc())
	site.Use(middleware.CasbinMiddleware())
	{
		site.POST("/add", h.AddNavSite)
		site.POST("/update", h.UpdateNavSite)
		site.POST("/delete", h.DeleteNavSite)
	}
	return r
}
