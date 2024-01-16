package routes

import (
	"micro-net-hub/internal/server/handler"

	"github.com/gin-gonic/gin"
)

func InitSiteNavRoutes(r *gin.RouterGroup) gin.IRoutes {
	api := r.Group("/sitenav")
	{
		var h handler.SiteNavHandler
		api.GET("/getall", h.GetNavSites)
	}

	return r
}
