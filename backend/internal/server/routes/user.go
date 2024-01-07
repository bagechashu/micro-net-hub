package routes

import (
	"micro-net-hub/internal/server/handler"
	"micro-net-hub/internal/server/middleware"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// 注册用户路由
func InitUserRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	user := r.Group("/user")
	// 开启jwt认证中间件
	user.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	user.Use(middleware.CasbinMiddleware())
	{
		user.GET("/info", handler.User.GetUserInfo)                   // 暂时未完成
		user.GET("/list", handler.User.List)                          // 用户列表
		user.POST("/add", handler.User.Add)                           // 添加用户
		user.POST("/update", handler.User.Update)                     // 更新用户
		user.POST("/delete", handler.User.Delete)                     // 删除用户
		user.POST("/changePwd", handler.User.ChangePwd)               // 修改用户密码
		user.POST("/changeUserStatus", handler.User.ChangeUserStatus) // 修改用户状态

		user.POST("/syncDingTalkUsers", handler.User.SyncDingTalkUsers) // 同步钉钉用户到平台
		user.POST("/syncWeComUsers", handler.User.SyncWeComUsers)       // 同步企业微信用户到平台
		user.POST("/syncFeiShuUsers", handler.User.SyncFeiShuUsers)     // 同步飞书用户到平台
		user.POST("/syncOpenLdapUsers", handler.User.SyncOpenLdapUsers) // 同步Ldap用户到平台
		user.POST("/syncSqlUsers", handler.User.SyncSqlUsers)           // 同步Sql用户到Ldap
	}
	return r
}
