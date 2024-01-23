package handler

import (
	"fmt"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/module/goldap/usermgr"
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/helper"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct{}

// List 记录列表
func (GroupHandler) List(c *gin.Context) {
	req := new(userModel.GroupListReq)
	helper.HandleRequest(c, req, userLogic.GroupLogicIns.List)
}

// UserInGroup 在分组内的用户
func (GroupHandler) UserInGroup(c *gin.Context) {
	req := new(userModel.UserInGroupReq)
	helper.HandleRequest(c, req, userLogic.GroupLogicIns.UserInGroup)
}

// UserNoInGroup 不在分组的用户
func (GroupHandler) UserNoInGroup(c *gin.Context) {
	req := new(userModel.UserNoInGroupReq)
	helper.HandleRequest(c, req, userLogic.GroupLogicIns.UserNoInGroup)
}

// GetTree 接口树
func (GroupHandler) GetTree(c *gin.Context) {
	req := new(userModel.GroupListReq)
	helper.HandleRequest(c, req, userLogic.GroupLogicIns.GetTree)
}

// Add 新建记录
func (GroupHandler) Add(c *gin.Context) {
	req := new(userModel.GroupAddReq)
	helper.HandleRequest(c, req, userLogic.GroupLogicIns.Add)
}

// Update 更新记录
func (GroupHandler) Update(c *gin.Context) {
	req := new(userModel.GroupUpdateReq)
	helper.HandleRequest(c, req, userLogic.GroupLogicIns.Update)
}

// Delete 删除记录
func (GroupHandler) Delete(c *gin.Context) {
	req := new(userModel.GroupDeleteReq)
	helper.HandleRequest(c, req, userLogic.GroupLogicIns.Delete)
}

// AddUser 添加用户
func (GroupHandler) AddUser(c *gin.Context) {
	req := new(userModel.GroupAddUserReq)
	helper.HandleRequest(c, req, userLogic.GroupLogicIns.AddUser)
}

// RemoveUser 移除用户
func (GroupHandler) RemoveUser(c *gin.Context) {
	req := new(userModel.GroupRemoveUserReq)
	helper.HandleRequest(c, req, userLogic.GroupLogicIns.RemoveUser)
}

// 同步钉钉部门信息
func (GroupHandler) SyncDingTalkDepts(c *gin.Context) {
	if config.Conf.DingTalk == nil {
		c.JSON(http.StatusOK, helper.NewConfigError(fmt.Errorf("没有 钉钉-DingTalk 相关配置")))
		return
	}
	req := new(userModel.SyncDingTalkDeptsReq)
	um := usermgr.NewDingTalk()
	helper.HandleRequest(c, req, um.SyncDepts)
}

// 同步企业微信部门信息
func (GroupHandler) SyncWeComDepts(c *gin.Context) {
	if config.Conf.WeCom == nil {
		c.JSON(http.StatusOK, helper.NewConfigError(fmt.Errorf("没有 企业微信-Wechat 相关配置")))
		return
	}
	req := new(userModel.SyncWeComDeptsReq)
	um := usermgr.NewWeChat()
	helper.HandleRequest(c, req, um.SyncDepts)
}

// 同步飞书部门信息
func (GroupHandler) SyncFeiShuDepts(c *gin.Context) {
	if config.Conf.FeiShu == nil {
		c.JSON(http.StatusOK, helper.NewConfigError(fmt.Errorf("没有 飞书-Feishu 相关配置")))
		return
	}
	req := new(userModel.SyncFeiShuDeptsReq)
	um := usermgr.NewFeiShu()
	helper.HandleRequest(c, req, um.SyncDepts)
}

// 同步原ldap部门信息
func (GroupHandler) SyncOpenLdapDepts(c *gin.Context) {
	req := new(userModel.SyncOpenLdapDeptsReq)
	um := usermgr.NewOpenLdap()
	helper.HandleRequest(c, req, um.SyncDepts)
}

// 同步Sql中的分组信息到ldap
func (GroupHandler) SyncSqlGroups(c *gin.Context) {
	req := new(userModel.SyncSqlGrooupsReq)
	helper.HandleRequest(c, req, userLogic.SqlLogicIns.SyncSqlGroups)
}
