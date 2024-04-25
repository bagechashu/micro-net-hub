package ldap

import (
	"fmt"
	"micro-net-hub/internal/config"
	accountModel "micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/module/goldap/ldapmgr"
	"micro-net-hub/internal/module/goldap/usermgr"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	"github.com/gin-gonic/gin"
)

// 同步钉钉用户信息
func SyncDingTalkUsers(c *gin.Context) {
	if config.Conf.DingTalk == nil {
		helper.Err(c, helper.NewConfigError(fmt.Errorf("没有 钉钉-DingTalk 相关配置")), nil)
		return
	}
	req := new(accountModel.SyncDingUserReq)
	um := usermgr.NewDingTalk()
	helper.HandleRequest(c, req, um.SyncUsers)
}

// 同步钉钉部门信息
func SyncDingTalkDepts(c *gin.Context) {
	if config.Conf.DingTalk == nil {
		helper.Err(c, helper.NewConfigError(fmt.Errorf("没有 钉钉-DingTalk 相关配置")), nil)
		return
	}
	req := new(accountModel.SyncDingTalkDeptsReq)
	um := usermgr.NewDingTalk()
	helper.HandleRequest(c, req, um.SyncDepts)
}

// 同步企业微信用户信息
func SyncWeComUsers(c *gin.Context) {
	if config.Conf.WeCom == nil {
		helper.Err(c, helper.NewConfigError(fmt.Errorf("没有 企业微信-Wechat 相关配置")), nil)
		return
	}
	req := new(accountModel.SyncWeComUserReq)
	um := usermgr.NewWeChat()
	helper.HandleRequest(c, req, um.SyncUsers)
}

// 同步企业微信部门信息
func SyncWeComDepts(c *gin.Context) {
	if config.Conf.WeCom == nil {
		helper.Err(c, helper.NewConfigError(fmt.Errorf("没有 企业微信-Wechat 相关配置")), nil)
		return
	}
	req := new(accountModel.SyncWeComDeptsReq)
	um := usermgr.NewWeChat()
	helper.HandleRequest(c, req, um.SyncDepts)
}

// 同步飞书用户信息
func SyncFeiShuUsers(c *gin.Context) {
	if config.Conf.FeiShu == nil {
		helper.Err(c, helper.NewConfigError(fmt.Errorf("没有 飞书-Feishu 相关配置")), nil)
		return
	}
	req := new(accountModel.SyncFeiShuUserReq)
	um := usermgr.NewFeiShu()
	helper.HandleRequest(c, req, um.SyncUsers)
}

// 同步飞书部门信息
func SyncFeiShuDepts(c *gin.Context) {
	if config.Conf.FeiShu == nil {
		helper.Err(c, helper.NewConfigError(fmt.Errorf("没有 飞书-Feishu 相关配置")), nil)
		return
	}
	req := new(accountModel.SyncFeiShuDeptsReq)
	um := usermgr.NewFeiShu()
	helper.HandleRequest(c, req, um.SyncDepts)
}

// 同步ldap用户信息
func SyncOpenLdapUsers(c *gin.Context) {
	req := new(accountModel.SyncOpenLdapUserReq)
	um := usermgr.NewOpenLdap()
	helper.HandleRequest(c, req, um.SyncUsers)
}

// 同步原ldap部门信息
func SyncOpenLdapDepts(c *gin.Context) {
	req := new(accountModel.SyncOpenLdapDeptsReq)
	um := usermgr.NewOpenLdap()
	helper.HandleRequest(c, req, um.SyncDepts)
}

// 同步sql用户信息到ldap
func SyncSqlUsers(c *gin.Context) {
	req := new(accountModel.SyncSqlUserReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.SyncSqlUserReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c
		// 1.获取所有用户
		for _, id := range r.UserIds {
			filter := tools.H{"id": int(id)}
			var u accountModel.User
			if !u.Exist(filter) {
				return nil, helper.NewMySqlError(fmt.Errorf("有用户不存在"))
			}
		}
		var users = accountModel.NewUsers()
		err := users.GetUserByIds(r.UserIds)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取用户信息失败: " + err.Error()))
		}
		// 2.再将用户添加到ldap
		for _, user := range users {
			err = ldapmgr.LdapUserAdd(user)
			if err != nil {
				return nil, helper.NewLdapError(fmt.Errorf("SyncUser向LDAP同步用户失败：" + err.Error()))
			}
			// 获取用户将要添加的分组
			var gs = accountModel.NewGroups()
			err := gs.GetGroupsByIds(tools.StringToSlice(user.DepartmentId, ","))
			if err != nil {
				return nil, helper.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
			}
			for _, group := range gs {
				//根据选择的部门，添加到部门内
				err = ldapmgr.LdapDeptAddUserToGroup(group.GroupDN, user.UserDN)
				if err != nil {
					return nil, helper.NewMySqlError(fmt.Errorf("向Ldap添加用户到分组关系失败：" + err.Error()))
				}
			}
			err = user.ChangeStatus(1)
			if err != nil {
				return nil, helper.NewLdapError(fmt.Errorf("用户同步完毕之后更新状态失败：" + err.Error()))
			}
		}

		return nil, nil
	})
}

// 同步Sql中的分组信息到ldap
func SyncSqlGroups(c *gin.Context) {
	req := new(accountModel.SyncSqlGrooupsReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.SyncSqlGrooupsReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c
		// 1.获取所有分组
		for _, id := range r.GroupIds {
			filter := tools.H{"id": int(id)}
			var g accountModel.Group
			if !g.Exist(filter) {
				return nil, helper.NewMySqlError(fmt.Errorf("有分组不存在"))
			}
		}
		var gs = accountModel.NewGroups()
		err := gs.GetGroupsByIds(r.GroupIds)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取分组信息失败: " + err.Error()))
		}
		// 2.再将分组添加到ldap
		for _, group := range gs {
			err = ldapmgr.LdapDeptAdd(group)
			if err != nil {
				return nil, helper.NewLdapError(fmt.Errorf("SyncUser向LDAP同步分组失败：" + err.Error()))
			}
			if len(group.Users) > 0 {
				for _, user := range group.Users {
					if user.UserDN == config.Conf.Ldap.AdminDN {
						continue
					}
					err = ldapmgr.LdapDeptAddUserToGroup(group.GroupDN, user.UserDN)
					if err != nil {
						return nil, helper.NewLdapError(fmt.Errorf("同步分组之后处理分组内的用户失败：" + err.Error()))
					}
				}
			}
			err = group.ChangeSyncState(1)
			if err != nil {
				return nil, helper.NewLdapError(fmt.Errorf("分组同步完毕之后更新状态失败：" + err.Error()))
			}
		}

		return nil, nil
	})
}
