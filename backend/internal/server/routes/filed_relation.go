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
		filed_relation.POST("/add", handler.FieldRelation.Add)
		filed_relation.GET("/list", handler.FieldRelation.List)
		filed_relation.POST("/update", handler.FieldRelation.Update)
		filed_relation.POST("/delete", handler.FieldRelation.Delete)
	}

	return r
}
