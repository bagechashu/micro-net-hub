package handler

import (
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
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.UserLogicIns.Add(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Update 更新记录
func (UserHandler) Update(c *gin.Context) {
	req := new(userModel.UserUpdateReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.UserLogicIns.Update(c, req)
	helper.HandleResponse(c, data, respErr)
}

// List 记录列表
func (UserHandler) List(c *gin.Context) {
	req := new(userModel.UserListReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.UserLogicIns.List(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Delete 删除记录
func (m UserHandler) Delete(c *gin.Context) {
	req := new(userModel.UserDeleteReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.UserLogicIns.Delete(c, req)
	helper.HandleResponse(c, data, respErr)
}

// ChangePwd 更新密码
func (m UserHandler) ChangePwd(c *gin.Context) {
	req := new(userModel.UserChangePwdReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.UserLogicIns.ChangePwd(c, req)
	helper.HandleResponse(c, data, respErr)
}

// ChangeUserStatus 更改用户状态
func (m UserHandler) ChangeUserStatus(c *gin.Context) {
	req := new(userModel.UserChangeUserStatusReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.UserLogicIns.ChangeUserStatus(c, req)
	helper.HandleResponse(c, data, respErr)
}

// GetUserInfo 获取当前登录用户信息
func (uc UserHandler) GetUserInfo(c *gin.Context) {
	req := new(userModel.UserInfoReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.UserLogicIns.GetUserInfo(c, req)
	helper.HandleResponse(c, data, respErr)
}

// 同步钉钉用户信息
func (uc UserHandler) SyncDingTalkUsers(c *gin.Context) {
	req := new(userModel.SyncDingUserReq)
	helper.BindAndValidateRequest(c, req)

	um := usermgr.NewDingTalk()
	data, respErr := um.SyncUsers(c, req)
	helper.HandleResponse(c, data, respErr)
}

// 同步企业微信用户信息
func (uc UserHandler) SyncWeComUsers(c *gin.Context) {
	req := new(userModel.SyncWeComUserReq)
	helper.BindAndValidateRequest(c, req)

	um := usermgr.NewWeChat()
	data, respErr := um.SyncUsers(c, req)
	helper.HandleResponse(c, data, respErr)
}

// 同步飞书用户信息
func (uc UserHandler) SyncFeiShuUsers(c *gin.Context) {
	req := new(userModel.SyncFeiShuUserReq)
	helper.BindAndValidateRequest(c, req)

	um := usermgr.NewFeiShu()
	data, respErr := um.SyncUsers(c, req)
	helper.HandleResponse(c, data, respErr)
}

// 同步ldap用户信息
func (uc UserHandler) SyncOpenLdapUsers(c *gin.Context) {
	req := new(userModel.SyncOpenLdapUserReq)
	helper.BindAndValidateRequest(c, req)

	um := usermgr.NewOpenLdap()
	data, respErr := um.SyncUsers(c, req)
	helper.HandleResponse(c, data, respErr)
}

// 同步sql用户信息到ldap
func (uc UserHandler) SyncSqlUsers(c *gin.Context) {
	req := new(userModel.SyncSqlUserReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.SqlLogicIns.SyncSqlUsers(c, req)
	helper.HandleResponse(c, data, respErr)
}
