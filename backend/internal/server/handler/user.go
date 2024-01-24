package handler

import (
	"fmt"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/module/goldap/usermgr"
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{}

// Add 添加记录
func (UserHandler) Add(c *gin.Context) {
	req := new(userModel.UserAddReq)
	helper.HandleRequest(c, req, userLogic.UserLogicIns.Add)
}

// Update 更新记录
func (UserHandler) Update(c *gin.Context) {
	req := new(userModel.UserUpdateReq)
	helper.HandleRequest(c, req, userLogic.UserLogicIns.Update)
}

// List 记录列表
func (UserHandler) List(c *gin.Context) {
	req := new(userModel.UserListReq)
	helper.HandleRequest(c, req, userLogic.UserLogicIns.List)
}

// Delete 删除记录
func (m UserHandler) Delete(c *gin.Context) {
	req := new(userModel.UserDeleteReq)
	helper.HandleRequest(c, req, userLogic.UserLogicIns.Delete)
}

// ChangePwd 更新密码
func (m UserHandler) ChangePwd(c *gin.Context) {
	req := new(userModel.UserChangePwdReq)
	helper.HandleRequest(c, req, userLogic.UserLogicIns.ChangePwd)
}

// ChangeUserStatus 更改用户状态
func (m UserHandler) ChangeUserStatus(c *gin.Context) {
	req := new(userModel.UserChangeUserStatusReq)
	helper.HandleRequest(c, req, userLogic.UserLogicIns.ChangeUserStatus)
}

// GetUserInfo 获取当前登录用户信息
func (uc UserHandler) GetUserInfo(c *gin.Context) {
	req := new(userModel.UserInfoReq)
	helper.HandleRequest(c, req, userLogic.UserLogicIns.GetUserInfo)
}

// 同步钉钉用户信息
func (uc UserHandler) SyncDingTalkUsers(c *gin.Context) {
	if config.Conf.DingTalk == nil {
		helper.Err(c, helper.NewConfigError(fmt.Errorf("没有 钉钉-DingTalk 相关配置")), nil)
		return
	}
	req := new(userModel.SyncDingUserReq)
	um := usermgr.NewDingTalk()
	helper.HandleRequest(c, req, um.SyncUsers)
}

// 同步企业微信用户信息
func (uc UserHandler) SyncWeComUsers(c *gin.Context) {
	if config.Conf.WeCom == nil {
		helper.Err(c, helper.NewConfigError(fmt.Errorf("没有 企业微信-Wechat 相关配置")), nil)
		return
	}
	req := new(userModel.SyncWeComUserReq)
	um := usermgr.NewWeChat()
	helper.HandleRequest(c, req, um.SyncUsers)
}

// 同步飞书用户信息
func (uc UserHandler) SyncFeiShuUsers(c *gin.Context) {
	if config.Conf.FeiShu == nil {
		helper.Err(c, helper.NewConfigError(fmt.Errorf("没有 飞书-Feishu 相关配置")), nil)
		return
	}
	req := new(userModel.SyncFeiShuUserReq)
	um := usermgr.NewFeiShu()
	helper.HandleRequest(c, req, um.SyncUsers)
}

// 同步ldap用户信息
func (uc UserHandler) SyncOpenLdapUsers(c *gin.Context) {
	req := new(userModel.SyncOpenLdapUserReq)
	um := usermgr.NewOpenLdap()
	helper.HandleRequest(c, req, um.SyncUsers)
}

// 同步sql用户信息到ldap
func (uc UserHandler) SyncSqlUsers(c *gin.Context) {
	req := new(userModel.SyncSqlUserReq)
	helper.HandleRequest(c, req, userLogic.SqlLogicIns.SyncSqlUsers)
}
