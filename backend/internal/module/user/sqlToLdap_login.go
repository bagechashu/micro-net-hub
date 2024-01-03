package user

import (
	"fmt"

	"micro-net-hub/internal/module/goldap/ldapmgr"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/config"
	"micro-net-hub/internal/tools"

	"github.com/gin-gonic/gin"
)

type SqlLogic struct{}

// SyncSqlUsers 同步sql的用户信息到ldap
func (d *SqlLogic) SyncSqlUsers(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.SyncSqlUserReq)
	if !ok {
		return nil, tools.ReqAssertErr
	}
	_ = c
	// 1.获取所有用户
	for _, id := range r.UserIds {
		filter := tools.H{"id": int(id)}
		if !userModel.UserSrvIns.Exist(filter) {
			return nil, tools.NewMySqlError(fmt.Errorf("有用户不存在"))
		}
	}
	users, err := userModel.UserSrvIns.GetUserByIds(r.UserIds)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取用户信息失败: " + err.Error()))
	}
	// 2.再将用户添加到ldap
	for _, user := range users {
		err = ldapmgr.LdapUserAdd(&user)
		if err != nil {
			return nil, tools.NewLdapError(fmt.Errorf("SyncUser向LDAP同步用户失败：" + err.Error()))
		}
		// 获取用户将要添加的分组
		groups, err := userModel.GroupSrvIns.GetGroupByIds(tools.StringToSlice(user.DepartmentId, ","))
		if err != nil {
			return nil, tools.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
		}
		for _, group := range groups {
			//根据选择的部门，添加到部门内
			err = ldapmgr.LdapDeptAddUserToGroup(group.GroupDN, user.UserDN)
			if err != nil {
				return nil, tools.NewMySqlError(fmt.Errorf("向Ldap添加用户到分组关系失败：" + err.Error()))
			}
		}
		err = userModel.UserSrvIns.ChangeSyncState(int(user.ID), 1)
		if err != nil {
			return nil, tools.NewLdapError(fmt.Errorf("用户同步完毕之后更新状态失败：" + err.Error()))
		}
	}

	return nil, nil
}

// SyncSqlGroups 同步sql中的分组信息到ldap
func (d *SqlLogic) SyncSqlGroups(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.SyncSqlGrooupsReq)
	if !ok {
		return nil, tools.ReqAssertErr
	}
	_ = c
	// 1.获取所有分组
	for _, id := range r.GroupIds {
		filter := tools.H{"id": int(id)}
		if !userModel.GroupSrvIns.Exist(filter) {
			return nil, tools.NewMySqlError(fmt.Errorf("有分组不存在"))
		}
	}
	groups, err := userModel.GroupSrvIns.GetGroupByIds(r.GroupIds)
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取分组信息失败: " + err.Error()))
	}
	// 2.再将分组添加到ldap
	for _, group := range groups {
		err = ldapmgr.LdapDeptAdd(group)
		if err != nil {
			return nil, tools.NewLdapError(fmt.Errorf("SyncUser向LDAP同步分组失败：" + err.Error()))
		}
		if len(group.Users) > 0 {
			for _, user := range group.Users {
				if user.UserDN == config.Conf.Ldap.AdminDN {
					continue
				}
				err = ldapmgr.LdapDeptAddUserToGroup(group.GroupDN, user.UserDN)
				if err != nil {
					return nil, tools.NewLdapError(fmt.Errorf("同步分组之后处理分组内的用户失败：" + err.Error()))
				}
			}
		}
		err = userModel.GroupSrvIns.ChangeSyncState(int(group.ID), 1)
		if err != nil {
			return nil, tools.NewLdapError(fmt.Errorf("分组同步完毕之后更新状态失败：" + err.Error()))
		}
	}

	return nil, nil
}

// SearchGroupDiff 检索未同步到ldap中的分组
func SearchGroupDiff() (err error) {
	// 获取sql中的数据
	var sqlGroupList []*userModel.Group
	sqlGroupList, err = userModel.GroupSrvIns.ListAll()
	if err != nil {
		return err
	}
	// 获取ldap中的数据
	var ldapGroupList []*userModel.Group
	ldapGroupList, err = ldapmgr.LdapDeptListGroupDN()
	if err != nil {
		return err
	}
	// 比对两个系统中的数据
	groups := diffGroup(sqlGroupList, ldapGroupList)
	for _, group := range groups {
		if group.GroupDN == config.Conf.Ldap.BaseDN {
			continue
		}
		err = userModel.GroupSrvIns.ChangeSyncState(int(group.ID), 2)
	}
	return
}

// SearchUserDiff 检索未同步到ldap中的用户
func SearchUserDiff() (err error) {
	// 获取sql中的数据
	var sqlUserList []*userModel.User
	sqlUserList, err = userModel.UserSrvIns.ListAll()
	if err != nil {
		return err
	}
	// 获取ldap中的数据
	var ldapUserList []*userModel.User
	ldapUserList, err = ldapmgr.LdapUserListUserDN()
	if err != nil {
		return err
	}
	// 比对两个系统中的数据
	users := diffUser(sqlUserList, ldapUserList)
	for _, user := range users {
		if user.UserDN == config.Conf.Ldap.AdminDN {
			continue
		}
		err = userModel.UserSrvIns.ChangeSyncState(int(user.ID), 2)
	}
	return
}

// diffGroup 比较出sql中有但ldap中没有的group列表
func diffGroup(sqlGroup, ldapGroup []*userModel.Group) (rst []*userModel.Group) {
	var tmp = make(map[string]struct{}, 0)

	for _, v := range ldapGroup {
		tmp[v.GroupDN] = struct{}{}
	}

	for _, v := range sqlGroup {
		if _, ok := tmp[v.GroupDN]; !ok {
			rst = append(rst, v)
		}
	}
	return
}

// diffUser 比较出sql中有但ldap中没有的user列表
func diffUser(sqlUser, ldapUser []*userModel.User) (rst []*userModel.User) {
	var tmp = make(map[string]struct{}, len(sqlUser))

	for _, v := range ldapUser {
		tmp[v.UserDN] = struct{}{}
	}

	for _, v := range sqlUser {
		if _, ok := tmp[v.UserDN]; !ok {
			rst = append(rst, v)
		}
	}
	return
}
