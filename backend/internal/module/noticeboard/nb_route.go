package noticeboard

import (
	"micro-net-hub/internal/module/account/role"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitNoticeMgrRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	noticeboard := r.Group("/notice")
	{
		noticeboard.GET("/getall", GetAll)
	}

	mgr := noticeboard.Group("/mgr")
	mgr.Use(authMiddleware.MiddlewareFunc())
	mgr.Use(role.CasbinMiddleware())
	{
		mgr.POST("/add", AddNoticeBoard)
		mgr.POST("/update", UpdateNoticeBoard)
		mgr.POST("/delete", DeleteNoticeBoard)
	}
	return r
}
