package routes

import (
	"micro-net-hub/internal/server/handler"
	"micro-net-hub/internal/server/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitGroupRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	group := r.Group("/group")
	// 开启jwt认证中间件
	group.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	group.Use(middleware.CasbinMiddleware())
	{
		group.GET("/list", handler.Group.List)
		group.GET("/tree", handler.Group.GetTree)
		group.POST("/add", handler.Group.Add)
		group.POST("/update", handler.Group.Update)
		group.POST("/delete", handler.Group.Delete)
		group.POST("/adduser", handler.Group.AddUser)
		group.POST("/removeuser", handler.Group.RemoveUser)

		group.GET("/useringroup", handler.Group.UserInGroup)
		group.GET("/usernoingroup", handler.Group.UserNoInGroup)

		group.POST("/syncDingTalkDepts", handler.Group.SyncDingTalkDepts) // 同步钉钉部门到平台
		group.POST("/syncWeComDepts", handler.Group.SyncWeComDepts)       // 同步企业微信部门到平台
		group.POST("/syncFeiShuDepts", handler.Group.SyncFeiShuDepts)     // 同步飞书部门到平台
		group.POST("/syncOpenLdapDepts", handler.Group.SyncOpenLdapDepts) // 同步ldap的分组到平台InitGroupRoutes
		group.POST("/syncSqlGroups", handler.Group.SyncSqlGroups)         // 同步Sql分组到Ldap
	}

	return r
}
