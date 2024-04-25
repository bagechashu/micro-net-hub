package user

import (
	accountModel "micro-net-hub/internal/module/account/model"
	dashboardModel "micro-net-hub/internal/module/dashboard/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"

	"fmt"

	"micro-net-hub/internal/config"
	apiMgrModel "micro-net-hub/internal/module/apimgr/model"
	"micro-net-hub/internal/module/goldap/ldapmgr"
	opLogModel "micro-net-hub/internal/module/operationlog/model"
	"micro-net-hub/internal/tools"
)

// SendCode 给用户邮箱发送验证码
func SendCode(c *gin.Context) {
	req := new(accountModel.BaseSendCodeReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.BaseSendCodeReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c
		// 判断邮箱是否正确
		user := new(accountModel.User)
		err := user.Find(tools.H{"mail": r.Mail})
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("通过邮箱查询用户失败" + err.Error()))
		}
		if user.Status != 1 || user.SyncState != 1 {
			return nil, helper.NewMySqlError(fmt.Errorf("该用户已离职或者未同步在ldap，无法重置密码，如有疑问，请联系管理员"))
		}

		if !config.Conf.Email.Enable {
			return nil, helper.NewValidatorError(fmt.Errorf("邮件通知功能未启用, 请联系管理员"))
		}
		// global.Log.Debugf("SendCode Request User: %+v", user)
		err = tools.SendVerificationCode([]string{r.Mail})
		if err != nil {
			return nil, helper.NewLdapError(fmt.Errorf("邮件发送验证码失败, 请联系管理员" + err.Error()))
		}

		return nil, nil
	})
}

// ChangePwd  忘记密码,用户通过邮箱的验证码重置密码
func ForgetPwd(c *gin.Context) {
	req := new(accountModel.BaseChangePwdReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.BaseChangePwdReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c
		// 判断邮箱是否正确
		var u accountModel.User
		if !u.Exist(tools.H{"mail": r.Mail}) {
			return nil, helper.NewValidatorError(fmt.Errorf("邮箱不存在,请检查邮箱是否正确"))
		}
		// 判断验证码是否过期
		cacheCode, ok := tools.VerificationCodeCache.Get(r.Mail)
		if !ok {
			return nil, helper.NewValidatorError(fmt.Errorf("对不起，该验证码已超过5分钟有效期，请重新重新密码"))
		}
		// 判断验证码是否正确
		if cacheCode != r.Code {
			return nil, helper.NewValidatorError(fmt.Errorf("验证码错误，请检查邮箱中正确的验证码，如果点击多次发送验证码，请用最后一次生成的验证码来验证"))
		}

		user := new(accountModel.User)
		err := user.Find(tools.H{"mail": r.Mail})
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("通过邮箱查询用户失败" + err.Error()))
		}

		newpass, err := ldapmgr.LdapUserNewPwd(user.Username)
		if err != nil {
			return nil, helper.NewLdapError(fmt.Errorf("LDAP生成新密码失败" + err.Error()))
		}

		// 更新数据库密码
		err = user.ChangePwd(tools.NewGenPasswd(newpass))
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("在MySQL更新密码失败: " + err.Error()))
		}

		if !config.Conf.Email.Enable {
			return nil, helper.NewValidatorError(fmt.Errorf("邮件通知功能未启用, 请联系管理员"))
		}
		err = tools.SendNewPass([]string{user.Mail}, newpass)
		if err != nil {
			return nil, helper.NewLdapError(fmt.Errorf("邮件发送新密码失败, 请联系管理员" + err.Error()))
		}
		return nil, nil
	})
}

// Dashboard 系统首页展示数据
func Dashboard(c *gin.Context) {
	req := new(dashboardModel.BaseDashboardReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		_, ok := req.(*dashboardModel.BaseDashboardReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		userCount, err := accountModel.UserCount()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取用户总数失败"))
		}

		groupCount, err := accountModel.GroupCount()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取分组总数失败"))
		}
		roleCount, err := accountModel.RoleCount()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取角色总数失败"))
		}
		menuCount, err := accountModel.MenuCount()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取菜单总数失败"))
		}
		apiCount, err := apiMgrModel.Count()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取接口总数失败"))
		}
		logCount, err := opLogModel.Count()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取日志总数失败"))
		}

		rst := make([]*dashboardModel.DashboardListRsp, 0)

		rst = append(rst,
			&dashboardModel.DashboardListRsp{
				DataType:  "user",
				DataName:  "用户",
				DataCount: userCount,
				Icon:      "people",
				Path:      "#/personnel/user",
			},
			&dashboardModel.DashboardListRsp{
				DataType:  "group",
				DataName:  "分组",
				DataCount: groupCount,
				Icon:      "peoples",
				Path:      "#/personnel/group",
			},
			&dashboardModel.DashboardListRsp{
				DataType:  "role",
				DataName:  "角色",
				DataCount: roleCount,
				Icon:      "eye-open",
				Path:      "#/system/role",
			},
			&dashboardModel.DashboardListRsp{
				DataType:  "menu",
				DataName:  "菜单",
				DataCount: menuCount,
				Icon:      "tree-table",
				Path:      "#/system/menu",
			},
			&dashboardModel.DashboardListRsp{
				DataType:  "api",
				DataName:  "接口",
				DataCount: apiCount,
				Icon:      "tree",
				Path:      "#/system/api",
			},
			&dashboardModel.DashboardListRsp{
				DataType:  "log",
				DataName:  "日志",
				DataCount: logCount,
				Icon:      "documentation",
				Path:      "#/log/operation-log",
			},
		)

		return rst, nil
	})
}

// EncryptPasswd 生成加密密码
func EncryptPasswd(c *gin.Context) {
	req := new(accountModel.EncryptPasswdReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.EncryptPasswdReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		return tools.NewGenPasswd(r.Passwd), nil
	})
}

// DecryptPasswd 密码解密为明文
func DecryptPasswd(c *gin.Context) {
	req := new(accountModel.DecryptPasswdReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.DecryptPasswdReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		return tools.NewParsePasswd(r.Passwd), nil
	})
}
