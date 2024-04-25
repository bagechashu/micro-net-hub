package user

import (
	"micro-net-hub/internal/module/account/role"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// 注册用户路由
func InitUserRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	user := r.Group("/user")
	// 开启jwt认证中间件
	user.Use(authMiddleware.MiddlewareFunc())
	// 开启casbin鉴权中间件
	user.Use(role.CasbinMiddleware())
	{
		user.GET("/info", GetUserInfo)                   // 暂时未完成
		user.GET("/list", List)                          // 用户列表
		user.POST("/add", Add)                           // 添加用户
		user.POST("/update", Update)                     // 更新用户
		user.POST("/delete", Delete)                     // 删除用户
		user.POST("/changePwd", ChangePwd)               // 修改用户密码
		user.POST("/resetTotpSecret", ReSetTotpSecret)   //重置 Totp Secret
		user.POST("/changeUserStatus", ChangeUserStatus) // 修改用户状态
	}
	return r
}
