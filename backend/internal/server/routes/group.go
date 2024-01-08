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
		var h handler.GroupHandler
		group.GET("/list", h.List)
		group.GET("/tree", h.GetTree)
		group.POST("/add", h.Add)
		group.POST("/update", h.Update)
		group.POST("/delete", h.Delete)
		group.POST("/adduser", h.AddUser)
		group.POST("/removeuser", h.RemoveUser)

		group.GET("/useringroup", h.UserInGroup)
		group.GET("/usernoingroup", h.UserNoInGroup)

		group.POST("/syncDingTalkDepts", h.SyncDingTalkDepts) // 同步钉钉部门到平台
		group.POST("/syncWeComDepts", h.SyncWeComDepts)       // 同步企业微信部门到平台
		group.POST("/syncFeiShuDepts", h.SyncFeiShuDepts)     // 同步飞书部门到平台
		group.POST("/syncOpenLdapDepts", h.SyncOpenLdapDepts) // 同步ldap的分组到平台InitGroupRoutes
		group.POST("/syncSqlGroups", h.SyncSqlGroups)         // 同步Sql分组到Ldap
	}

	return r
}
