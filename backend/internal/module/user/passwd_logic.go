package user

import (
	"fmt"

	apiMgrModel "micro-net-hub/internal/module/apimgr/model"
	dashboardModel "micro-net-hub/internal/module/dashboard/model"
	"micro-net-hub/internal/module/goldap/ldapmgr"
	opLogModel "micro-net-hub/internal/module/operation_log/model"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	"github.com/gin-gonic/gin"
)

type PasswdLogic struct{}

// SendCode 发送验证码
func (l PasswdLogic) SendCode(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.BaseSendCodeReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c
	// 判断邮箱是否正确
	user := new(userModel.User)
	err := user.Find(tools.H{"mail": r.Mail})
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("通过邮箱查询用户失败" + err.Error()))
	}
	if user.Status != 1 || user.SyncState != 1 {
		return nil, helper.NewMySqlError(fmt.Errorf("该用户已离职或者未同步在ldap，无法重置密码，如有疑问，请联系管理员"))
	}
	err = tools.SendCode([]string{r.Mail})
	if err != nil {
		return nil, helper.NewLdapError(fmt.Errorf("邮件发送失败" + err.Error()))
	}

	return nil, nil
}

// ChangePwd 重置密码
func (l PasswdLogic) ChangePwd(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.BaseChangePwdReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c
	// 判断邮箱是否正确
	var u userModel.User
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

	user := new(userModel.User)
	err := user.Find(tools.H{"mail": r.Mail})
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("通过邮箱查询用户失败" + err.Error()))
	}

	newpass, err := ldapmgr.LdapUserNewPwd(user.Username)
	if err != nil {
		return nil, helper.NewLdapError(fmt.Errorf("LDAP生成新密码失败" + err.Error()))
	}

	err = tools.SendMail([]string{user.Mail}, newpass)
	if err != nil {
		return nil, helper.NewLdapError(fmt.Errorf("邮件发送失败" + err.Error()))
	}

	// 更新数据库密码
	err = user.ChangePwd(tools.NewGenPasswd(newpass))
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("在MySQL更新密码失败: " + err.Error()))
	}

	return nil, nil
}

// Dashboard 仪表盘
func (l PasswdLogic) Dashboard(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	_, ok := req.(*dashboardModel.BaseDashboardReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	userCount, err := userModel.UserCount()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取用户总数失败"))
	}

	groupCount, err := userModel.GroupCount()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取分组总数失败"))
	}
	roleCount, err := userModel.RoleCount()
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取角色总数失败"))
	}
	menuCount, err := userModel.MenuCount()
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
}

// EncryptPasswd
func (l PasswdLogic) EncryptPasswd(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.EncryptPasswdReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	return tools.NewGenPasswd(r.Passwd), nil
}

// DecryptPasswd
func (l PasswdLogic) DecryptPasswd(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.DecryptPasswdReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	return tools.NewParPasswd(r.Passwd), nil
}
