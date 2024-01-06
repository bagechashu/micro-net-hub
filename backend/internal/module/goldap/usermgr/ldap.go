package usermgr

import (
	"fmt"

	"micro-net-hub/internal/module/goldap/ldapmgr"
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"

	"micro-net-hub/internal/tools"

	"github.com/gin-gonic/gin"
)

type OpenLdap struct{}

func NewOpenLdap() OpenLdap {
	return OpenLdap{}
}

// 通过ldap获取部门信息
func (mgr OpenLdap) SyncDepts(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	// 1.获取所有部门
	depts, err := ldapmgr.LdapDeptGetAll()
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("获取ldap部门列表失败：%s", err.Error()))
	}
	groups := make([]*userModel.Group, 0)
	for _, dept := range depts {
		groups = append(groups, &userModel.Group{
			GroupName:          dept.Name,
			Remark:             dept.Remark,
			SourceDeptId:       dept.Id,
			SourceDeptParentId: dept.ParentId,
			GroupDN:            dept.DN,
		})
	}
	// 2.将远程数据转换成树
	deptTree := userLogic.GroupListToTree("0", groups)

	// 3.根据树进行创建
	err = ldapmgr.LdapDeptAddRec(deptTree.Children)

	return nil, err
}

// 根据现有数据库同步到的部门信息，开启用户同步
func (mgr OpenLdap) SyncUsers(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	// 1.获取ldap用户列表
	staffs, err := ldapmgr.LDAPUserGetAll()
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("获取ldap用户列表失败：%s", err.Error()))
	}
	// 2.遍历用户，开始写入
	for _, staff := range staffs {
		groupIds, err := userModel.DeptIdsToGroupIds(staff.DepartmentIds)
		if err != nil {
			return nil, tools.NewMySqlError(fmt.Errorf("将部门ids转换为内部部门id失败：%s", err.Error()))
		}
		// 根据角色id获取角色
		roles := userModel.NewRoles()
		err = roles.GetRolesByIds([]uint{2})
		if err != nil {
			return nil, tools.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败:%s", err.Error()))
		}
		// 入库
		err = ldapmgr.LdapUserAdd(&userModel.User{
			Username:      staff.Name,
			Nickname:      staff.DisplayName,
			GivenName:     staff.GivenName,
			Mail:          staff.Mail,
			JobNumber:     staff.EmployeeNumber,
			Mobile:        staff.Mobile,
			PostalAddress: staff.PostalAddress,
			Departments:   staff.BusinessCategory,
			Position:      staff.DepartmentNumber,
			Introduction:  staff.CN,
			Creator:       "system",
			Source:        "openldap",
			DepartmentId:  tools.SliceToString(groupIds, ","),
			SourceUserId:  staff.Name,
			SourceUnionId: staff.Name,
			Roles:         roles,
			UserDN:        staff.DN,
		})
		if err != nil {
			return nil, tools.NewOperationError(fmt.Errorf("SyncOpenLdapUsers写入用户失败：%s", err.Error()))
		}
	}
	return nil, nil
}
