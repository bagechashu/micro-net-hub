package routes

import (
	"micro-net-hub/internal/server/handler"

	"github.com/gin-gonic/gin"
)

func InitSiteNavRoutes(r *gin.RouterGroup) gin.IRoutes {
	api := r.Group("/sitenav")
	{
		var h handler.SiteNavHandler
		api.GET("/getall", h.GetAllNavConfig)
		api.GET("/getsidegroups", h.GetSideNavGroups)
		api.GET("/getgroups", h.GetNavGroups)
	}

	return r
}
