package controller

import (
	"micro-net-hub/internal/module/goldap/usermgr"
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"

	"github.com/gin-gonic/gin"
)

type GroupController struct{}

// List 记录列表
func (m *GroupController) List(c *gin.Context) {
	req := new(userModel.GroupListReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.GroupLogicIns.List(c, req)
	})
}

// UserInGroup 在分组内的用户
func (m *GroupController) UserInGroup(c *gin.Context) {
	req := new(userModel.UserInGroupReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.GroupLogicIns.UserInGroup(c, req)
	})
}

// UserNoInGroup 不在分组的用户
func (m *GroupController) UserNoInGroup(c *gin.Context) {
	req := new(userModel.UserNoInGroupReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.GroupLogicIns.UserNoInGroup(c, req)
	})
}

// GetTree 接口树
func (m *GroupController) GetTree(c *gin.Context) {
	req := new(userModel.GroupListReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.GroupLogicIns.GetTree(c, req)
	})
}

// Add 新建记录
func (m *GroupController) Add(c *gin.Context) {
	req := new(userModel.GroupAddReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.GroupLogicIns.Add(c, req)
	})
}

// Update 更新记录
func (m *GroupController) Update(c *gin.Context) {
	req := new(userModel.GroupUpdateReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.GroupLogicIns.Update(c, req)
	})
}

// Delete 删除记录
func (m *GroupController) Delete(c *gin.Context) {
	req := new(userModel.GroupDeleteReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.GroupLogicIns.Delete(c, req)
	})
}

// AddUser 添加用户
func (m *GroupController) AddUser(c *gin.Context) {
	req := new(userModel.GroupAddUserReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.GroupLogicIns.AddUser(c, req)
	})
}

// RemoveUser 移除用户
func (m *GroupController) RemoveUser(c *gin.Context) {
	req := new(userModel.GroupRemoveUserReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.GroupLogicIns.RemoveUser(c, req)
	})
}

// 同步钉钉部门信息
func (m *GroupController) SyncDingTalkDepts(c *gin.Context) {
	req := new(userModel.SyncDingTalkDeptsReq)
	Run(c, req, func() (interface{}, interface{}) {
		um := usermgr.NewDingTalk()
		return um.SyncDepts(c, req)
	})
}

// 同步企业微信部门信息
func (m *GroupController) SyncWeComDepts(c *gin.Context) {
	req := new(userModel.SyncWeComDeptsReq)
	Run(c, req, func() (interface{}, interface{}) {
		um := usermgr.NewWeChat()
		return um.SyncDepts(c, req)
	})
}

// 同步飞书部门信息
func (m *GroupController) SyncFeiShuDepts(c *gin.Context) {
	req := new(userModel.SyncFeiShuDeptsReq)
	Run(c, req, func() (interface{}, interface{}) {
		um := usermgr.NewFeiShu()
		return um.SyncDepts(c, req)
	})
}

// 同步原ldap部门信息
func (m *GroupController) SyncOpenLdapDepts(c *gin.Context) {
	req := new(userModel.SyncOpenLdapDeptsReq)
	Run(c, req, func() (interface{}, interface{}) {
		um := usermgr.NewOpenLdap()
		return um.SyncDepts(c, req)
	})
}

// 同步Sql中的分组信息到ldap
func (m *GroupController) SyncSqlGroups(c *gin.Context) {
	req := new(userModel.SyncSqlGrooupsReq)
	Run(c, req, func() (interface{}, interface{}) {
		return userLogic.SqlLogicIns.SyncSqlGroups(c, req)
	})
}
