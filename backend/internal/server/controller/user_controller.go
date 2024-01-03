package controller

import (
	"micro-net-hub/internal/module/goldap/usermgr"
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"

	"github.com/gin-gonic/gin"
)

type UserController struct{}

// Add 添加记录
func (m *UserController) Add(c *gin.Context) {
	req := new(userModel.UserAddReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.UserLogicIns.Add(c, req)
	})
}

// Update 更新记录
func (m *UserController) Update(c *gin.Context) {
	req := new(userModel.UserUpdateReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.UserLogicIns.Update(c, req)
	})
}

// List 记录列表
func (m *UserController) List(c *gin.Context) {
	req := new(userModel.UserListReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.UserLogicIns.List(c, req)
	})
}

// Delete 删除记录
func (m UserController) Delete(c *gin.Context) {
	req := new(userModel.UserDeleteReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.UserLogicIns.Delete(c, req)
	})
}

// ChangePwd 更新密码
func (m UserController) ChangePwd(c *gin.Context) {
	req := new(userModel.UserChangePwdReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.UserLogicIns.ChangePwd(c, req)
	})
}

// ChangeUserStatus 更改用户状态
func (m UserController) ChangeUserStatus(c *gin.Context) {
	req := new(userModel.UserChangeUserStatusReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.UserLogicIns.ChangeUserStatus(c, req)
	})
}

// GetUserInfo 获取当前登录用户信息
func (uc UserController) GetUserInfo(c *gin.Context) {
	req := new(userModel.UserGetUserInfoReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.UserLogicIns.GetUserInfo(c, req)
	})
}

// 同步钉钉用户信息
func (uc UserController) SyncDingTalkUsers(c *gin.Context) {
	req := new(userModel.SyncDingUserReq)
	Run(c, req, func() (interface{}, interface{}) {
		um := usermgr.NewDingTalk()
		return um.SyncUsers(c, req)
	})
}

// 同步企业微信用户信息
func (uc UserController) SyncWeComUsers(c *gin.Context) {
	req := new(userModel.SyncWeComUserReq)
	Run(c, req, func() (interface{}, interface{}) {
		um := usermgr.NewWeChat()
		return um.SyncUsers(c, req)
	})
}

// 同步飞书用户信息
func (uc UserController) SyncFeiShuUsers(c *gin.Context) {
	req := new(userModel.SyncFeiShuUserReq)
	Run(c, req, func() (interface{}, interface{}) {
		um := usermgr.NewFeiShu()
		return um.SyncUsers(c, req)
	})
}

// 同步ldap用户信息
func (uc UserController) SyncOpenLdapUsers(c *gin.Context) {
	req := new(userModel.SyncOpenLdapUserReq)
	Run(c, req, func() (interface{}, interface{}) {
		um := usermgr.NewOpenLdap()
		return um.SyncUsers(c, req)
	})
}

// 同步sql用户信息到ldap
func (uc UserController) SyncSqlUsers(c *gin.Context) {
	req := new(userModel.SyncSqlUserReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.SqlLogicIns.SyncSqlUsers(c, req)
	})
}
