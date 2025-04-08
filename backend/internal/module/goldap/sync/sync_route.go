package sync

import (
	"micro-net-hub/internal/module/account/role"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitGoldapSyncRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	goldap := r.Group("/goldap/sync")
	// 开启jwt认证中间件
	goldap.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	goldap.Use(role.CasbinMiddleware())
	{
		// TODO: 修改成 get 请求
		goldap.POST("/syncDingTalkUsers", SyncDingTalkUsers) // 同步钉钉用户到平台
		goldap.POST("/syncDingTalkDepts", SyncDingTalkDepts) // 同步钉钉部门到平台
		goldap.POST("/syncWeComUsers", SyncWeComUsers)       // 同步企业微信用户到平台
		goldap.POST("/syncWeComDepts", SyncWeComDepts)       // 同步企业微信部门到平台
		goldap.POST("/syncFeiShuUsers", SyncFeiShuUsers)     // 同步飞书用户到平台
		goldap.POST("/syncFeiShuDepts", SyncFeiShuDepts)     // 同步飞书部门到平台
		goldap.POST("/syncOpenLdapDepts", SyncOpenLdapDepts) // 同步ldap的分组到平台
		goldap.POST("/syncOpenLdapUsers", SyncOpenLdapUsers) // 同步Ldap用户到平台
		goldap.POST("/syncSqlUsers", SyncSqlUsers)           // 同步Sql用户到Ldap
		goldap.POST("/syncSqlGroups", SyncSqlGroups)         // 同步Sql分组到Ldap
	}

	return r
}
