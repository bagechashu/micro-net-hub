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
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.PasswdLogicIns.SendCode(c, req)
	helper.HandleResponse(c, data, respErr)
}

// ChangePwd 用户通过邮箱修改密码
func (BaseHandler) ChangePwd(c *gin.Context) {
	req := new(userModel.BaseChangePwdReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.PasswdLogicIns.ChangePwd(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Dashboard 系统首页展示数据
func (BaseHandler) Dashboard(c *gin.Context) {
	req := new(dashboardModel.BaseDashboardReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.PasswdLogicIns.Dashboard(c, req)
	helper.HandleResponse(c, data, respErr)
}

// EncryptPasswd 生成加密密码
func (BaseHandler) EncryptPasswd(c *gin.Context) {
	req := new(userModel.EncryptPasswdReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.PasswdLogicIns.EncryptPasswd(c, req)
	helper.HandleResponse(c, data, respErr)
}

// DecryptPasswd 密码解密为明文
func (BaseHandler) DecryptPasswd(c *gin.Context) {
	req := new(userModel.DecryptPasswdReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.PasswdLogicIns.DecryptPasswd(c, req)
	helper.HandleResponse(c, data, respErr)
}
