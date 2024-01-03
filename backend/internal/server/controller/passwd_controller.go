package controller

import (
	dashboardModel "micro-net-hub/internal/module/dashboard/model"
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"

	"github.com/gin-gonic/gin"
)

type BaseController struct{}

// SendCode 给用户邮箱发送验证码
func (m *BaseController) SendCode(c *gin.Context) {
	req := new(userModel.BaseSendCodeReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.PasswdLogicIns.SendCode(c, req)
	})
}

// ChangePwd 用户通过邮箱修改密码
func (m *BaseController) ChangePwd(c *gin.Context) {
	req := new(userModel.BaseChangePwdReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.PasswdLogicIns.ChangePwd(c, req)
	})
}

// Dashboard 系统首页展示数据
func (m *BaseController) Dashboard(c *gin.Context) {
	req := new(dashboardModel.BaseDashboardReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.PasswdLogicIns.Dashboard(c, req)
	})
}

// EncryptPasswd 生成加密密码
func (m *BaseController) EncryptPasswd(c *gin.Context) {
	req := new(userModel.EncryptPasswdReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.PasswdLogicIns.EncryptPasswd(c, req)
	})
}

// DecryptPasswd 密码解密为明文
func (m *BaseController) DecryptPasswd(c *gin.Context) {
	req := new(userModel.DecryptPasswdReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.PasswdLogicIns.DecryptPasswd(c, req)
	})
}
