package routes

import (
	"micro-net-hub/internal/server/handler"
	"micro-net-hub/internal/server/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitFieldRelationRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	filed_relation := r.Group("/fieldrelation")
	// 开启jwt认证中间件
	filed_relation.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	filed_relation.Use(middleware.CasbinMiddleware())
	{
		var h handler.FieldRelationHandler
		filed_relation.POST("/add", h.Add)
		filed_relation.GET("/list", h.List)
		filed_relation.POST("/update", h.Update)
		filed_relation.POST("/delete", h.Delete)
	}

	return r
}
