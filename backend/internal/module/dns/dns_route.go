package dns

import (
	"micro-net-hub/internal/module/account/role"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitDnsMgrRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	sitenav := r.Group("/dns")
	{
		sitenav.GET("/getall", GetAll)
	}

	group := sitenav.Group("/zone")
	group.Use(authMiddleware.MiddlewareFunc())
	group.Use(role.CasbinMiddleware())
	{
		group.POST("/add", AddZone)
		group.POST("/update", UpdateZone)
		group.POST("/delete", DeleteZone)
	}

	site := sitenav.Group("/record")
	site.Use(authMiddleware.MiddlewareFunc())
	site.Use(role.CasbinMiddleware())
	{
		site.POST("/add", AddRecord)
		site.POST("/update", UpdateRecord)
		site.POST("/delete", DeleteRecord)
	}
	return r
}
