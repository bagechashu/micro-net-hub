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

type GroupHandler struct{}

// List 记录列表
func (GroupHandler) List(c *gin.Context) {
	req := new(userModel.GroupListReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.GroupLogicIns.List(c, req)
	helper.HandleResponse(c, data, respErr)
}

// UserInGroup 在分组内的用户
func (GroupHandler) UserInGroup(c *gin.Context) {
	req := new(userModel.UserInGroupReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.GroupLogicIns.UserInGroup(c, req)
	helper.HandleResponse(c, data, respErr)
}

// UserNoInGroup 不在分组的用户
func (GroupHandler) UserNoInGroup(c *gin.Context) {
	req := new(userModel.UserNoInGroupReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.GroupLogicIns.UserNoInGroup(c, req)
	helper.HandleResponse(c, data, respErr)
}

// GetTree 接口树
func (GroupHandler) GetTree(c *gin.Context) {
	req := new(userModel.GroupListReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.GroupLogicIns.GetTree(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Add 新建记录
func (GroupHandler) Add(c *gin.Context) {
	req := new(userModel.GroupAddReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.GroupLogicIns.Add(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Update 更新记录
func (GroupHandler) Update(c *gin.Context) {
	req := new(userModel.GroupUpdateReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.GroupLogicIns.Update(c, req)
	helper.HandleResponse(c, data, respErr)
}

// Delete 删除记录
func (GroupHandler) Delete(c *gin.Context) {
	req := new(userModel.GroupDeleteReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.GroupLogicIns.Delete(c, req)
	helper.HandleResponse(c, data, respErr)
}

// AddUser 添加用户
func (GroupHandler) AddUser(c *gin.Context) {
	req := new(userModel.GroupAddUserReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.GroupLogicIns.AddUser(c, req)
	helper.HandleResponse(c, data, respErr)
}

// RemoveUser 移除用户
func (GroupHandler) RemoveUser(c *gin.Context) {
	req := new(userModel.GroupRemoveUserReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.GroupLogicIns.RemoveUser(c, req)
	helper.HandleResponse(c, data, respErr)
}

// 同步钉钉部门信息
func (GroupHandler) SyncDingTalkDepts(c *gin.Context) {
	req := new(userModel.SyncDingTalkDeptsReq)
	helper.BindAndValidateRequest(c, req)

	if config.Conf.DingTalk == nil {
		helper.HandleResponse(c, nil, helper.NewConfigError(fmt.Errorf("没有 钉钉-Dingtalk 相关配置")))
		return
	}
	um := usermgr.NewDingTalk()
	data, respErr := um.SyncDepts(c, req)
	helper.HandleResponse(c, data, respErr)
}

// 同步企业微信部门信息
func (GroupHandler) SyncWeComDepts(c *gin.Context) {
	req := new(userModel.SyncWeComDeptsReq)
	helper.BindAndValidateRequest(c, req)

	if config.Conf.WeCom == nil {
		helper.HandleResponse(c, nil, helper.NewConfigError(fmt.Errorf("没有 企业微信-Wechat 相关配置")))
		return
	}
	um := usermgr.NewWeChat()
	data, respErr := um.SyncDepts(c, req)
	helper.HandleResponse(c, data, respErr)
}

// 同步飞书部门信息
func (GroupHandler) SyncFeiShuDepts(c *gin.Context) {
	req := new(userModel.SyncFeiShuDeptsReq)
	helper.BindAndValidateRequest(c, req)

	if config.Conf.FeiShu == nil {
		helper.HandleResponse(c, nil, helper.NewConfigError(fmt.Errorf("没有 飞书-Feishu 相关配置")))
		return
	}
	um := usermgr.NewFeiShu()
	data, respErr := um.SyncDepts(c, req)
	helper.HandleResponse(c, data, respErr)
}

// 同步原ldap部门信息
func (GroupHandler) SyncOpenLdapDepts(c *gin.Context) {
	req := new(userModel.SyncOpenLdapDeptsReq)
	helper.BindAndValidateRequest(c, req)

	um := usermgr.NewOpenLdap()
	data, respErr := um.SyncDepts(c, req)
	helper.HandleResponse(c, data, respErr)
}

// 同步Sql中的分组信息到ldap
func (GroupHandler) SyncSqlGroups(c *gin.Context) {
	req := new(userModel.SyncSqlGrooupsReq)
	helper.BindAndValidateRequest(c, req)

	data, respErr := userLogic.SqlLogicIns.SyncSqlGroups(c, req)
	helper.HandleResponse(c, data, respErr)
}
