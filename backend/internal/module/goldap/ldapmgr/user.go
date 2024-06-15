package ldapmgr

import (
	"fmt"
	"strings"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	accountModel "micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	ldap "github.com/go-ldap/ldap/v3"
)

type LdapUser struct {
	Name             string   `json:"name"`
	DN               string   `json:"dn"`
	CN               string   `json:"cn"`
	SN               string   `json:"sn"`
	Mobile           string   `json:"mobile"`
	BusinessCategory string   `json:"businessCategory"` // 业务类别，部门名字
	DepartmentNumber string   `json:"departmentNumber"` // 部门编号，此处可以存放员工的职位
	Description      string   `json:"description"`      // 描述
	DisplayName      string   `json:"displayName"`      // 展示名字，可以是中文名字
	Mail             string   `json:"mail"`             // 邮箱
	EmployeeNumber   string   `json:"employeeNumber"`   // 员工工号
	GivenName        string   `json:"givenName"`        // 给定名字，如果公司有花名，可以用这个字段
	PostalAddress    string   `json:"postalAddress"`    // 家庭住址
	DepartmentDns    []string `json:"department_dns"`
}

// 创建资源
func LdapUserAdd(user *accountModel.User) error {
	user.CheckAttrVacancies()
	add := ldap.NewAddRequest(user.UserDN, nil)
	add.Attribute("objectClass", []string{"inetOrgPerson"})
	add.Attribute("uid", []string{user.Username})
	add.Attribute("cn", []string{user.Username})
	add.Attribute("sn", []string{user.Nickname})
	add.Attribute("givenName", []string{user.GivenName})
	add.Attribute("description", []string{user.Introduction})
	add.Attribute("displayName", []string{user.Nickname})
	add.Attribute("mail", []string{user.Mail})
	add.Attribute("userPassword", []string{tools.EncodePass([]byte(tools.NewParsePasswd(user.Password)))})
	add.Attribute("employeeNumber", []string{user.JobNumber})
	// add.Attribute("businessCategory", []string{user.Departments})
	add.Attribute("departmentNumber", []string{user.Position})
	add.Attribute("postalAddress", []string{user.PostalAddress})
	add.Attribute("mobile", []string{user.Mobile})

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return err
	}

	return conn.Add(add)
}

// Update 更新资源
func LdapUserUpdate(oldusername string, user *accountModel.User) error {
	modify := ldap.NewModifyRequest(user.UserDN, nil)
	modify.Replace("cn", []string{user.Username})
	modify.Replace("sn", []string{oldusername})
	// modify.Replace("businessCategory", []string{user.Departments})
	modify.Replace("departmentNumber", []string{user.Position})
	modify.Replace("description", []string{user.Introduction})
	modify.Replace("displayName", []string{user.Nickname})
	modify.Replace("mail", []string{user.Mail})
	modify.Replace("employeeNumber", []string{user.JobNumber})
	modify.Replace("givenName", []string{user.GivenName})
	modify.Replace("postalAddress", []string{user.PostalAddress})
	modify.Replace("mobile", []string{user.Mobile})

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return err
	}

	err = conn.Modify(modify)
	if err != nil {
		return err
	}
	if config.Conf.Ldap.UserNameModify && oldusername != user.Username {
		modifyDn := ldap.NewModifyDNRequest(fmt.Sprintf("uid=%s,%s", oldusername, config.Conf.Ldap.UserDN), fmt.Sprintf("uid=%s", user.Username), true, "")
		return conn.ModifyDN(modifyDn)
	}
	return nil
}

func LdapUserExist(filter map[string]interface{}) (bool, error) {
	filter_str := ""
	for key, value := range filter {
		filter_str += fmt.Sprintf("(%s=%s)", key, value)
	}
	search_filter := fmt.Sprintf("(&(|(objectClass=inetOrgPerson)(objectClass=simpleSecurityObject))%s)", filter_str)
	// Construct query request
	searchRequest := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,                                     // This is basedn, we will start searching from this node.
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, // Here several parameters are respectively scope, derefAliases, sizeLimit, timeLimit,  typesOnly
		search_filter,  // This is Filter for LDAP query
		[]string{"DN"}, // Here are the attributes returned by the query, provided as an array. If empty, all attributes are returned
		nil,
	)

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return false, err
	}
	var sr *ldap.SearchResult
	// Search through ldap built-in search
	sr, err = conn.Search(searchRequest)
	if err != nil {
		return false, err
	}
	if len(sr.Entries) > 0 {
		return true, nil
	}
	return false, nil
}

// Delete 删除资源
func LdapUserDelete(udn string) error {
	del := ldap.NewDelRequest(udn, nil)
	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return err
	}
	return conn.Del(del)
}

// ChangePwd 修改用户密码，此处旧密码也可以为空，ldap可以直接通过用户DN加上新密码来进行修改
func LdapUserChangePwd(udn, oldpasswd, newpasswd string) error {
	modifyPass := ldap.NewPasswordModifyRequest(udn, oldpasswd, newpasswd)

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return err
	}

	_, err = conn.PasswordModify(modifyPass)
	if err != nil {
		return fmt.Errorf("password modify failed for %s, err: %v", udn, err)
	}
	return nil
}

// NewPwd 新旧密码都是空，通过管理员可以修改成功并返回新的密码
func LdapUserNewPwd(username string) (string, error) {
	udn := fmt.Sprintf("uid=%s,%s", username, config.Conf.Ldap.UserDN)
	if username == "admin" {
		udn = config.Conf.Ldap.AdminDN
	}
	modifyPass := ldap.NewPasswordModifyRequest(udn, "", "")

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return "", err
	}

	newpass, err := conn.PasswordModify(modifyPass)
	if err != nil {
		return "", fmt.Errorf("password modify failed for %s, err: %v", username, err)
	}
	return newpass.GeneratedPassword, nil
}
func LdapUserListUserDN() (users []*accountModel.User, err error) {
	// Construct query request
	searchRequest := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,                                     // This is basedn, we will start searching from this node.
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, // Here several parameters are respectively scope, derefAliases, sizeLimit, timeLimit,  typesOnly
		"(|(objectClass=inetOrgPerson)(objectClass=simpleSecurityObject))", // This is Filter for LDAP query
		[]string{"DN"}, // Here are the attributes returned by the query, provided as an array. If empty, all attributes are returned
		nil,
	)

	// 获取 LDAP 连接
	conn, err := global.LdapPool.GetConn()
	defer global.LdapPool.PutConn(conn)
	if err != nil {
		return users, err
	}
	var sr *ldap.SearchResult
	// Search through ldap built-in search
	sr, err = conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	if len(sr.Entries) > 0 {
		for _, v := range sr.Entries {
			users = append(users, &accountModel.User{
				UserDN: v.DN,
			})
		}
	}
	return
}

// GetAllUsers 获取所有员工信息 V2
func LdapUserGetAllV2() (ret []*LdapUser, err error) {
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
			if v.DN == config.Conf.Ldap.UserDN || !strings.Contains(v.DN, config.Conf.Ldap.UserDN) {
				continue
			}
			name := strings.Split(strings.Split(v.DN, ",")[0], "=")[1]
			// 获取部门DN
			deptDns, err := LdapUserGetDeptDns(v.DN)
			if err != nil {
				return ret, err
			}
			ret = append(ret, &LdapUser{
				Name:             name,
				DN:               v.DN,
				CN:               v.GetAttributeValue("cn"),
				SN:               v.GetAttributeValue("sn"),
				Mobile:           v.GetAttributeValue("mobile"),
				BusinessCategory: v.GetAttributeValue("businessCategory"),
				DepartmentNumber: v.GetAttributeValue("departmentNumber"),
				Description:      v.GetAttributeValue("description"),
				DisplayName:      v.GetAttributeValue("displayName"),
				Mail:             v.GetAttributeValue("mail"),
				EmployeeNumber:   v.GetAttributeValue("employeeNumber"),
				GivenName:        v.GetAttributeValue("givenName"),
				PostalAddress:    v.GetAttributeValue("postalAddress"),
				DepartmentDns:    deptDns,
			})
		}
	}
	return
}

// GetUserDeptDns 获取用户所在的部门 DN
func LdapUserGetDeptDns(udn string) (groupDns []string, err error) {
	// Construct query request
	searchRequest := ldap.NewSearchRequest(
		config.Conf.Ldap.BaseDN,                                     // This is basedn, we will start searching from this node.
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false, // Here several parameters are respectively scope, derefAliases, sizeLimit, timeLimit,  typesOnly
		fmt.Sprintf("(|(Member=%s)(uniqueMember=%s))", udn, udn), // This is Filter for LDAP query
		[]string{}, // Here are the attributes returned by the query, provided as an array. If empty, all attributes are returned
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
		return nil, err
	}
	for _, v := range sr.Entries {
		groupDns = append(groupDns, v.DN)
	}

	return groupDns, nil
}

// 添加 Ldap 用户数据 到数据库
func LdapUserSyncToDB(user *accountModel.User) error {
	// 根据 user_dn 查询用户,不存在则创建
	if !user.Exist(tools.H{"user_dn": user.UserDN}) {
		user.CheckAttrVacancies()
		// 先将用户添加到MySQL
		err := user.Add()
		if err != nil {
			return helper.NewMySqlError(fmt.Errorf("向MySQL创建用户失败：" + err.Error()))
		}

		// 获取用户将要添加的分组

		var gs = accountModel.NewGroups()
		err = gs.GetGroupsByIds(tools.StringToSlice(user.DepartmentIds, ","))
		if err != nil {
			return helper.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
		}
		for _, group := range gs {
			if group.GroupDN[:3] == "ou=" {
				continue
			}
			// 先将用户和部门信息维护到MySQL
			err := group.AddUserToGroup(user)
			if err != nil {
				return helper.NewMySqlError(fmt.Errorf("向MySQL添加用户到分组关系失败：" + err.Error()))
			}
		}
		return nil
	}
	return nil
}
