package handler

import (
	dashboardModel "micro-net-hub/internal/module/dashboard/model"
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type BaseHandler struct{}

// SendCode 给用户邮箱发送验证码
func (BaseHandler) SendCode(c *gin.Context) {
	req := new(userModel.BaseSendCodeReq)
	helper.HandleRequest(c, req, userLogic.PasswdLogicIns.SendCode)
}

// ChangePwd 用户通过邮箱修改密码
func (BaseHandler) ChangePwd(c *gin.Context) {
	req := new(userModel.BaseChangePwdReq)
	helper.HandleRequest(c, req, userLogic.PasswdLogicIns.ChangePwd)
}

// Dashboard 系统首页展示数据
func (BaseHandler) Dashboard(c *gin.Context) {
	req := new(dashboardModel.BaseDashboardReq)
	helper.HandleRequest(c, req, userLogic.PasswdLogicIns.Dashboard)
}

// EncryptPasswd 生成加密密码
func (BaseHandler) EncryptPasswd(c *gin.Context) {
	req := new(userModel.EncryptPasswdReq)
	helper.HandleRequest(c, req, userLogic.PasswdLogicIns.EncryptPasswd)
}

// DecryptPasswd 密码解密为明文
func (BaseHandler) DecryptPasswd(c *gin.Context) {
	req := new(userModel.DecryptPasswdReq)
	helper.HandleRequest(c, req, userLogic.PasswdLogicIns.DecryptPasswd)
}
