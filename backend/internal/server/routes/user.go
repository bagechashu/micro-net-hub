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
		var h handler.UserHandler
		user.GET("/info", h.GetUserInfo)                   // 暂时未完成
		user.GET("/list", h.List)                          // 用户列表
		user.POST("/add", h.Add)                           // 添加用户
		user.POST("/update", h.Update)                     // 更新用户
		user.POST("/delete", h.Delete)                     // 删除用户
		user.POST("/changePwd", h.ChangePwd)               // 修改用户密码
		user.GET("/resetTotpSecret", h.ReSetTotpSecret)    //重置 Totp Secret
		user.POST("/changeUserStatus", h.ChangeUserStatus) // 修改用户状态

		user.POST("/syncDingTalkUsers", h.SyncDingTalkUsers) // 同步钉钉用户到平台
		user.POST("/syncWeComUsers", h.SyncWeComUsers)       // 同步企业微信用户到平台
		user.POST("/syncFeiShuUsers", h.SyncFeiShuUsers)     // 同步飞书用户到平台
		user.POST("/syncOpenLdapUsers", h.SyncOpenLdapUsers) // 同步Ldap用户到平台
		user.POST("/syncSqlUsers", h.SyncSqlUsers)           // 同步Sql用户到Ldap
	}
	return r
}
