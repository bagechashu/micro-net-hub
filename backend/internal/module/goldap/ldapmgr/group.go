package ldapmgr

import (
	"errors"
	"fmt"
	"strings"

	"micro-net-hub/internal/global"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/config"
	"micro-net-hub/internal/tools"

	ldap "github.com/go-ldap/ldap/v3"
)

type LdapDept struct {
	DN       string `json:"dn"`
	Id       string `json:"id"`       // 部门ID
	Name     string `json:"name"`     // 部门名称拼音
	Remark   string `json:"remark"`   // 部门中文名
	ParentId string `json:"parentid"` // 父部门ID
}

// Add 添加资源
func LdapDeptAdd(g *userModel.Group) error { //organizationalUnit
	if g.Remark == "" {
		g.Remark = g.GroupName
	}
	add := ldap.NewAddRequest(g.GroupDN, nil)
	if g.GroupType == "ou" {
		add.Attribute("objectClass", []string{"organizationalUnit", "top"}) // 如果定义了 groupOfNAmes，那么必须指定member，否则报错如下：object class 'groupOfNames' requires attribute 'member'
	}
	if g.GroupType == "cn" {
		add.Attribute("objectClass", []string{"groupOfUniqueNames", "top"})
		add.Attribute("uniqueMember", []string{config.Conf.Ldap.AdminDN}) // 所以这里创建组的时候，默认将admin加入其中，以免创建时没有人员而报上边的错误
	}
	add.Attribute(g.GroupType, []string{g.GroupName})
	add.Attribute("description", []string{g.Remark})

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return err
	}

	return conn.Add(add)
}

// AddGroup 添加部门数据
func LdapDeptsAdd(group *userModel.Group) error {
	// 判断部门名称是否存在,此处使用ldap中的唯一值dn,以免出现数据同步不全的问题
	if !userModel.GroupSrvIns.Exist(tools.H{"group_dn": group.GroupDN}) {
		// 此时的 group 已经附带了Build后动态关联好的字段，接下来将一些确定性的其他字段值添加上，就可以创建这个分组了
		group.Creator = "system"
		group.GroupType = strings.Split(strings.Split(group.GroupDN, ",")[0], "=")[0]
		parentid, err := LdapDeptGetParentGroupID(group)
		if err != nil {
			return err
		}
		group.ParentId = parentid
		group.Source = "openldap"
		err = userModel.GroupSrvIns.Add(group)
		if err != nil {
			return err
		}
	}
	return nil
}

// 添加部门
func LdapDeptAddRec(depts []*userModel.Group) error {
	for _, dept := range depts {
		err := LdapDeptsAdd(dept)
		if err != nil {
			return tools.NewOperationError(fmt.Errorf("DsyncOpenLdapDepts添加部门失败: %s", err.Error()))
		}
		if len(dept.Children) != 0 {
			err = LdapDeptAddRec(dept.Children)
			if err != nil {
				return tools.NewOperationError(fmt.Errorf("DsyncOpenLdapDepts添加部门失败: %s", err.Error()))
			}
		}
	}
	return nil
}

// UpdateGroup 更新一个分组
func LdapDeptUpdate(oldGroup, newGroup *userModel.Group) error {
	modify1 := ldap.NewModifyRequest(oldGroup.GroupDN, nil)
	modify1.Replace("description", []string{newGroup.Remark})

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return err
	}

	err = conn.Modify(modify1)
	if err != nil {
		return err
	}
	// 如果配置文件允许修改分组名称，且分组名称发生了变化，那么执行修改分组名称
	if config.Conf.Ldap.GroupNameModify && newGroup.GroupName != oldGroup.GroupName {
		modify2 := ldap.NewModifyDNRequest(oldGroup.GroupDN, newGroup.GroupDN, true, "")
		err := conn.ModifyDN(modify2)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete 删除资源
func LdapDeptDelete(gdn string) error {
	del := ldap.NewDelRequest(gdn, nil)

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return err
	}

	return conn.Del(del)
}

// AddUserToGroup 添加用户到分组
func LdapDeptAddUserToGroup(dn, udn string) error {
	//判断dn是否以ou开头
	if dn[:3] == "ou=" {
		return errors.New("不能添加用户到OU组织单元")
	}
	newmr := ldap.NewModifyRequest(dn, nil)
	newmr.Add("uniqueMember", []string{udn})

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return err
	}

	return conn.Modify(newmr)
}

// DelUserFromGroup 将用户从分组删除
func LdapDeptRemoveUserFromGroup(gdn, udn string) error {
	newmr := ldap.NewModifyRequest(gdn, nil)
	newmr.Delete("uniqueMember", []string{udn})

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return err
	}

	return conn.Modify(newmr)
}

// DelUserFromGroup 将用户从分组删除
func LdapDeptListGroupDN() (groups []*userModel.Group, err error) {
	// Construct query request
	searchRequest := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,                                     // This is basedn, we will start searching from this node.
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, // Here several parameters are respectively scope, derefAliases, sizeLimit, timeLimit,  typesOnly
		"(|(objectClass=organizationalUnit)(objectClass=groupOfUniqueNames))", // This is Filter for LDAP query
		[]string{"DN"}, // Here are the attributes returned by the query, provided as an array. If empty, all attributes are returned
		nil,
	)

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return groups, err
	}
	var sr *ldap.SearchResult
	// Search through ldap built-in search
	sr, err = conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	if len(sr.Entries) > 0 {
		for _, v := range sr.Entries {
			groups = append(groups, &userModel.Group{
				GroupDN: v.DN,
			})
		}
	}
	return
}

// GetAllDepts 获取所有部门
func LdapDeptGetAll() (ret []*LdapDept, err error) {
	// Construct query request
	searchRequest := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,                                     // This is basedn, we will start searching from this node.
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, // Here several parameters are respectively scope, derefAliases, sizeLimit, timeLimit,  typesOnly
		"(&(objectClass=*))", // This is Filter for LDAP query
		[]string{},           // Here are the attributes returned by the query, provided as an array. If empty, all attributes are returned
		nil,
	)

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return nil, err
	}

	// Search through ldap built-in search
	sr, err := conn.Search(searchRequest)
	if err != nil {
		return ret, err
	}
	// Refers to the entry that returns data. If it is greater than 0, the interface returns normally.
	if len(sr.Entries) > 0 {
		for _, v := range sr.Entries {
			if v.DN == config.Conf.Ldap.BaseDN || v.DN == config.Conf.Ldap.AdminDN || strings.Contains(v.DN, config.Conf.Ldap.UserDN) {
				continue
			}
			var ele LdapDept
			ele.DN = v.DN
			ele.Name = strings.Split(strings.Split(v.DN, ",")[0], "=")[1]
			ele.Id = strings.Split(strings.Split(v.DN, ",")[0], "=")[1]
			ele.Remark = v.GetAttributeValue("description")
			if len(strings.Split(v.DN, ","))-len(strings.Split(config.Conf.Ldap.BaseDN, ",")) == 1 {
				ele.ParentId = "0"
			} else {
				ele.ParentId = strings.Split(strings.Split(v.DN, ",")[1], "=")[1]
			}
			ret = append(ret, &ele)
		}
	}
	return
}

// AddGroup 添加部门数据
func LdapDeptGetParentGroupID(group *userModel.Group) (id uint, err error) {
	switch group.SourceDeptParentId {
	case "dingtalkroot":
		group.SourceDeptParentId = "dingtalk_1"
	case "feishuroot":
		group.SourceDeptParentId = "feishu_0"
	case "wecomroot":
		group.SourceDeptParentId = "wecom_1"
	}
	parentGroup := new(userModel.Group)
	err = userModel.GroupSrvIns.Find(tools.H{"source_dept_id": group.SourceDeptParentId}, parentGroup)
	if err != nil {
		return id, tools.NewMySqlError(fmt.Errorf("查询父级部门失败：%s,%s", err.Error(), group.GroupName))
	}
	return parentGroup.ID, nil
}
