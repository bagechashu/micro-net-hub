package fieldrelation

import (
	"micro-net-hub/internal/module/account/role"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitFieldRelationRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	filed_relation := r.Group("/fieldrelation")
	// 开启jwt认证中间件
	filed_relation.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	filed_relation.Use(role.CasbinMiddleware())
	{
		filed_relation.POST("/add", Add)
		filed_relation.GET("/list", List)
		filed_relation.POST("/update", Update)
		filed_relation.POST("/delete", Delete)
	}

	return r
}
