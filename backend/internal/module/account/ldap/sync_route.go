package ldap

import (
	"micro-net-hub/internal/module/account/role"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

func InitGroupRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	gldap := r.Group("/gldap")
	// 开启jwt认证中间件
	gldap.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	gldap.Use(role.CasbinMiddleware())
	{
		gldap.POST("/syncDingTalkUsers", SyncDingTalkUsers) // 同步钉钉用户到平台
		gldap.POST("/syncDingTalkDepts", SyncDingTalkDepts) // 同步钉钉部门到平台
		gldap.POST("/syncWeComUsers", SyncWeComUsers)       // 同步企业微信用户到平台
		gldap.POST("/syncWeComDepts", SyncWeComDepts)       // 同步企业微信部门到平台
		gldap.POST("/syncFeiShuUsers", SyncFeiShuUsers)     // 同步飞书用户到平台
		gldap.POST("/syncFeiShuDepts", SyncFeiShuDepts)     // 同步飞书部门到平台
		gldap.POST("/syncOpenLdapDepts", SyncOpenLdapDepts) // 同步ldap的分组到平台
		gldap.POST("/syncOpenLdapUsers", SyncOpenLdapUsers) // 同步Ldap用户到平台
		gldap.POST("/syncSqlUsers", SyncSqlUsers)           // 同步Sql用户到Ldap
		gldap.POST("/syncSqlGroups", SyncSqlGroups)         // 同步Sql分组到Ldap
	}

	return r
}
