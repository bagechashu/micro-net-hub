package usermgr

import (
	"fmt"

	accountModel "micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/module/account/user"
	"micro-net-hub/internal/module/goldap/ldapmgr"
	"micro-net-hub/internal/server/helper"
)

type OpenLdap struct{}

func NewOpenLdap() OpenLdap {
	return OpenLdap{}
}

// 同步 ldap部门信息 到 数据库
func (mgr OpenLdap) SyncDepts() error {
	// 1.获取所有部门
	depts, err := ldapmgr.LdapDeptGetAll()
	if err != nil {
		return helper.NewOperationError(fmt.Errorf("获取ldap部门列表失败：%s", err.Error()))
	}
	groups := make([]*accountModel.Group, 0)
	for _, dept := range depts {
		groups = append(groups, &accountModel.Group{
			GroupName:          dept.Name,
			Remark:             dept.Remark,
			SourceDeptId:       dept.Id,
			SourceDeptParentId: dept.ParentId,
			GroupDN:            dept.DN,
		})
	}
	// 2.将远程数据转换成树
	deptTree := user.GroupListToTree("0", groups)

	// 3.根据树进行创建
	return ldapmgr.LdapDeptsSyncToDBRec(deptTree.Children)
}

// 同步 ldap用户信息 到 数据库
func (mgr OpenLdap) SyncUsers() error {
	// 1.获取ldap用户列表
	staffs, err := ldapmgr.LdapUserGetAllV2()
	if err != nil {
		return helper.NewOperationError(fmt.Errorf("获取ldap用户列表失败：%s", err.Error()))
	}
	// 2.遍历用户，开始写入
	for _, staff := range staffs {
		gs, err := accountModel.LdapDeptDnsToGroups(staff.DepartmentDns)
		if err != nil {
			return helper.NewMySqlError(fmt.Errorf("将部门ids转换为内部部门id失败：%s", err.Error()))
		}
		// 根据角色id获取角色
		roles := accountModel.NewRoles()
		err = roles.GetRolesByIds([]uint{2})
		if err != nil {
			return helper.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败:%s", err.Error()))
		}
		// 入库
		err = ldapmgr.LdapUserSyncToDB(&accountModel.User{
			Username:      staff.Name,
			Nickname:      staff.DisplayName,
			GivenName:     staff.GivenName,
			Mail:          staff.Mail,
			JobNumber:     staff.EmployeeNumber,
			Mobile:        staff.Mobile,
			PostalAddress: staff.PostalAddress,
			Position:      staff.DepartmentNumber,
			Introduction:  staff.CN,
			Creator:       "system",
			Source:        "openldap",
			SourceUserId:  staff.Name,
			SourceUnionId: staff.Name,
			Roles:         roles,
			UserDN:        staff.DN,
			Groups:        gs,
		})
		if err != nil {
			return helper.NewOperationError(fmt.Errorf("SyncOpenLdapUsers写入用户失败：%s", err.Error()))
		}
	}
	return nil
}
