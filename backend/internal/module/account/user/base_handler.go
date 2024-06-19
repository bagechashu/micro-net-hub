package user

import (
	accountModel "micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"

	"fmt"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/module/goldap/ldapmgr"
	"micro-net-hub/internal/tools"
)

// BaseSendCodeReq 发送验证码
type BaseSendCodeReq struct {
	Mail string `json:"mail" validate:"required,min=0,max=100"`
}

// SendCode 给用户邮箱发送验证码
func SendCode(c *gin.Context) {
	var req BaseSendCodeReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 判断邮箱是否正确
	user := new(accountModel.User)
	err = user.Find(map[string]interface{}{"mail": req.Mail})
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("通过邮箱查询用户失败"+err.Error())))
		return
	}
	if user.Status != 1 || user.SyncState != 1 {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("该用户已离职或者未同步在ldap，无法重置密码，如有疑问，请联系管理员")))
		return
	}

	if !config.Conf.Email.Enable {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("邮件通知功能未启用, 请联系管理员")))
		return
	}
	// global.Log.Debugf("SendCode Request User: %+v", user)
	err = tools.SendVerificationCode([]string{req.Mail})
	if err != nil {
		helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("邮件发送验证码失败, 请联系管理员"+err.Error())))
		return
	}

	helper.Success(c, nil)
}

// BaseChangePwdReq 修改密码结构体
type BaseChangePwdReq struct {
	Mail string `json:"mail" validate:"required,min=0,max=100"`
	Code string `json:"code" validate:"required,len=6"`
}

// ChangePwd  忘记密码,用户通过邮箱的验证码重置密码
func ForgetPwd(c *gin.Context) {
	var req BaseChangePwdReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 判断邮箱是否正确
	var u accountModel.User
	if !u.Exist(map[string]interface{}{"mail": req.Mail}) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("邮箱不存在,请检查邮箱是否正确")))
		return
	}
	// 判断验证码是否过期
	cacheCode, ok := tools.VerificationCodeCache.Get(req.Mail)
	if !ok {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("对不起，该验证码已超过5分钟有效期，请重新重新密码")))
		return
	}
	// 判断验证码是否正确
	if cacheCode != req.Code {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("验证码错误，请检查邮箱中正确的验证码，如果点击多次发送验证码，请用最后一次生成的验证码来验证")))
		return
	}

	user := new(accountModel.User)
	err = user.Find(map[string]interface{}{"mail": req.Mail})
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("通过邮箱查询用户失败"+err.Error())))
		return
	}

	newpass, err := ldapmgr.LdapUserNewPwd(user.Username)
	if err != nil {
		helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("LDAP生成新密码失败"+err.Error())))
		return
	}

	// 更新数据库密码
	err = user.ChangePwd(tools.NewGenPasswd(newpass))
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("在MySQL更新密码失败: "+err.Error())))
		return
	}

	if !config.Conf.Email.Enable {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("邮件通知功能未启用, 请联系管理员")))
		return
	}
	err = tools.SendNewPass([]string{user.Mail}, newpass)
	if err != nil {
		helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("邮件发送新密码失败, 请联系管理员"+err.Error())))
		return
	}

	helper.Success(c, nil)
}

// EncryptPasswdReq
type EncryptPasswdReq struct {
	Passwd string `json:"passwd" form:"passwd" validate:"required"`
}

// EncryptPasswd 生成加密密码
func EncryptPasswd(c *gin.Context) {
	var req EncryptPasswdReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}
	helper.Success(c, tools.NewGenPasswd(req.Passwd))
}

// DecryptPasswdReq
type DecryptPasswdReq struct {
	Passwd string `json:"passwd" form:"passwd" validate:"required"`
}

// DecryptPasswd 密码解密为明文
func DecryptPasswd(c *gin.Context) {
	var req DecryptPasswdReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}
	helper.Success(c, tools.NewParsePasswd(req.Passwd))
}
