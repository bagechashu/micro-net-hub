package user

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// 注册基础路由
func InitBaseRoutes(r *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) gin.IRoutes {
	base := r.Group("/base")
	{
		base.GET("/encryptpwd", EncryptPasswd) // 生成加密密码
		// base.GET("/decryptpwd", DecryptPasswd) // 密码解密为明文
		// 登录登出刷新token无需鉴权
		base.POST("/login", authMiddleware.LoginHandler)
		base.POST("/logout", authMiddleware.LogoutHandler)
		base.POST("/refreshToken", authMiddleware.RefreshHandler)
		base.POST("/sendcode", SendCode)   // 给用户邮箱发送验证码
		base.POST("/changePwd", ChangePwd) // 修改用户密码
	}
	return r
}
