package user

import (
	"fmt"
	"math/rand"
	"time"

	"micro-net-hub/internal/global"
	fieldRelationModel "micro-net-hub/internal/module/goldap/field_relation/model"
	"micro-net-hub/internal/module/goldap/ldapmgr"

	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/config"
	"micro-net-hub/internal/tools"

	"github.com/tidwall/gjson"
)

// CommonAddGroup 标准创建分组
func CommonAddGroup(group *userModel.Group) error {
	// 先在ldap中创建组
	err := ldapmgr.LdapDeptAdd(group)
	if err != nil {
		return err
	}

	// 然后在数据库中创建组
	err = group.Add()
	if err != nil {
		return err
	}

	// 默认创建分组之后，需要将admin添加到分组中
	adminInfo := new(userModel.User)
	err = userModel.UserSrvIns.Find(tools.H{"id": 1}, adminInfo)
	if err != nil {
		return err
	}

	err = group.AddUserToGroup([]userModel.User{*adminInfo})
	if err != nil {
		return err
	}

	return nil
}

// CommonUpdateGroup 标准更新分组
func CommonUpdateGroup(oldGroup, newGroup *userModel.Group) error {
	//若配置了不允许修改分组名称，则不更新分组名称
	if !config.Conf.Ldap.GroupNameModify {
		newGroup.GroupName = oldGroup.GroupName
	}

	err := ldapmgr.LdapDeptUpdate(oldGroup, newGroup)
	if err != nil {
		return err
	}
	err = newGroup.Update()
	if err != nil {
		return err
	}
	return nil
}

// CommonAddUser 标准创建用户
func CommonAddUser(user *userModel.User, groups []*userModel.Group) error {
	// 用户信息的预置处理
	if user.Nickname == "" {
		user.Nickname = "佚名"
	}
	if user.GivenName == "" {
		user.GivenName = user.Nickname
	}
	if user.Introduction == "" {
		user.Introduction = user.Nickname
	}
	if user.Mail == "" {
		// 兼容
		if len(config.Conf.Ldap.DefaultEmailSuffix) > 0 {
			user.Mail = user.Username + "@" + config.Conf.Ldap.DefaultEmailSuffix
		} else {
			user.Mail = user.Username + "@example.com"
		}
	}
	if user.JobNumber == "" {
		user.JobNumber = "0000"
	}
	if user.Departments == "" {
		user.Departments = "默认:研发中心"
	}
	if user.Position == "" {
		user.Position = "默认:打工人"
	}
	if user.PostalAddress == "" {
		user.PostalAddress = "默认:地球"
	}
	if user.Mobile == "" {
		user.Mobile = generateMobile()
	}

	// 先将用户添加到MySQL
	err := userModel.UserSrvIns.Add(user)
	if err != nil {
		return tools.NewMySqlError(fmt.Errorf("向MySQL创建用户失败：" + err.Error()))
	}
	// 再将用户添加到ldap
	err = ldapmgr.LdapUserAdd(user)
	if err != nil {
		return tools.NewLdapError(fmt.Errorf("AddUser向LDAP创建用户失败：" + err.Error()))
	}

	// 处理用户归属的组
	for _, group := range groups {
		if group.GroupDN[:3] == "ou=" {
			continue
		}
		// 先将用户和部门信息维护到MySQL
		err := group.AddUserToGroup([]userModel.User{*user})
		if err != nil {
			return tools.NewMySqlError(fmt.Errorf("向MySQL添加用户到分组关系失败：" + err.Error()))
		}
		//根据选择的部门，添加到部门内
		err = ldapmgr.LdapDeptAddUserToGroup(group.GroupDN, user.UserDN)
		if err != nil {
			return tools.NewMySqlError(fmt.Errorf("向Ldap添加用户到分组关系失败：" + err.Error()))
		}
	}
	return nil
}

// CommonUpdateUser 标准更新用户
func CommonUpdateUser(oldUser, newUser *userModel.User, groupId []uint) error {
	// 更新用户
	if !config.Conf.Ldap.UserNameModify {
		newUser.Username = oldUser.Username
	}

	err := ldapmgr.LdapUserUpdate(oldUser.Username, newUser)
	if err != nil {
		return tools.NewLdapError(fmt.Errorf("在LDAP更新用户失败：" + err.Error()))
	}

	err = userModel.UserSrvIns.Update(newUser)
	if err != nil {
		return tools.NewMySqlError(fmt.Errorf("在MySQL更新用户失败：" + err.Error()))
	}

	//判断部门信息是否有变化有变化则更新相应的数据库
	oldDeptIds := tools.StringToSlice(oldUser.DepartmentId, ",")
	addDeptIds, removeDeptIds := tools.ArrUintCmp(oldDeptIds, groupId)

	// 先处理添加的部门
	var addGroups = userModel.NewGroups()
	err = addGroups.GetGroupsByIds(addDeptIds)
	if err != nil {
		return tools.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
	}
	for _, group := range addGroups {
		if group.GroupDN[:3] == "ou=" {
			continue
		}
		// 先将用户和部门信息维护到MySQL
		err := group.AddUserToGroup([]userModel.User{*newUser})
		if err != nil {
			return tools.NewMySqlError(fmt.Errorf("向MySQL添加用户到分组关系失败：" + err.Error()))
		}
		//根据选择的部门，添加到部门内
		err = ldapmgr.LdapDeptAddUserToGroup(group.GroupDN, newUser.UserDN)
		if err != nil {
			return tools.NewLdapError(fmt.Errorf("向Ldap添加用户到分组关系失败：" + err.Error()))
		}
	}

	// 再处理删除的部门
	var removeGroups = userModel.NewGroups()
	err = removeGroups.GetGroupsByIds(removeDeptIds)
	if err != nil {
		return tools.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
	}
	for _, group := range removeGroups {
		if group.GroupDN[:3] == "ou=" {
			continue
		}
		err := group.RemoveUserFromGroup([]userModel.User{*newUser})
		if err != nil {
			return tools.NewMySqlError(fmt.Errorf("在MySQL将用户从分组移除失败：" + err.Error()))
		}
		err = ldapmgr.LdapDeptRemoveUserFromGroup(group.GroupDN, newUser.UserDN)
		if err != nil {
			return tools.NewMySqlError(fmt.Errorf("在ldap将用户从分组移除失败：" + err.Error()))
		}
	}
	return nil
}

// BuildGroupData 根据数据与动态字段组装成分组数据
func BuildGroupData(flag string, remoteData map[string]interface{}) (*userModel.Group, error) {
	output, err := json.Marshal(&remoteData)
	if err != nil {
		return nil, err
	}

	oldData := new(fieldRelationModel.FieldRelation)
	err = fieldRelationModel.Find(tools.H{"flag": flag + "_group"}, oldData)
	if err != nil {
		return nil, tools.NewMySqlError(err)
	}
	frs, err := tools.JsonToMap(string(oldData.Attributes))
	if err != nil {
		return nil, tools.NewOperationError(err)
	}

	g := &userModel.Group{}
	for system, remote := range frs {
		switch system {
		case "groupName":
			g.SetGroupName(gjson.Get(string(output), remote).String())
		case "remark":
			g.SetRemark(gjson.Get(string(output), remote).String())
		case "sourceDeptId":
			g.SetSourceDeptId(fmt.Sprintf("%s_%s", flag, gjson.Get(string(output), remote).String()))
		case "sourceDeptParentId":
			g.SetSourceDeptParentId(fmt.Sprintf("%s_%s", flag, gjson.Get(string(output), remote).String()))
		}
	}
	return g, nil
}

// BuildUserData 根据数据与动态字段组装成用户数据
func BuildUserData(flag string, remoteData map[string]interface{}) (*userModel.User, error) {
	output, err := json.Marshal(&remoteData)
	if err != nil {
		return nil, err
	}

	fieldRelationSource := new(fieldRelationModel.FieldRelation)
	err = fieldRelationModel.Find(tools.H{"flag": flag + "_user"}, fieldRelationSource)
	if err != nil {
		return nil, tools.NewMySqlError(err)
	}
	fieldRelation, err := tools.JsonToMap(string(fieldRelationSource.Attributes))
	if err != nil {
		return nil, tools.NewOperationError(err)
	}

	// 校验username是否为空，username为必填项
	name := gjson.Get(string(output), fieldRelation["username"]).String()
	if len(name) == 0 {
		global.Log.Warnf("%s 该用户未填写username", output)
		return nil, nil
	}

	u := &userModel.User{}
	for system, remote := range fieldRelation {
		switch system {
		case "username":
			u.SetUserName(gjson.Get(string(output), remote).String())
		case "nickname":
			u.SetNickName(gjson.Get(string(output), remote).String())
		case "givenName":
			u.SetGivenName(gjson.Get(string(output), remote).String())
		case "mail":
			u.SetMail(gjson.Get(string(output), remote).String())
		case "jobNumber":
			u.SetJobNumber(gjson.Get(string(output), remote).String())
		case "mobile":
			u.SetMobile(gjson.Get(string(output), remote).String())
		case "avatar":
			u.SetAvatar(gjson.Get(string(output), remote).String())
		case "postalAddress":
			u.SetPostalAddress(gjson.Get(string(output), remote).String())
		case "position":
			u.SetPosition(gjson.Get(string(output), remote).String())
		case "introduction":
			u.SetIntroduction(gjson.Get(string(output), remote).String())
		case "sourceUserId":
			u.SetSourceUserId(fmt.Sprintf("%s_%s", flag, gjson.Get(string(output), remote).String()))
		case "sourceUnionId":
			u.SetSourceUnionId(fmt.Sprintf("%s_%s", flag, gjson.Get(string(output), remote).String()))
		}
	}
	return u, nil
}

// ConvertDeptData 将部门信息转成本地结构体
func ConvertDeptData(flag string, remoteData []map[string]interface{}) (groups []*userModel.Group, err error) {
	for _, dept := range remoteData {
		group, err := BuildGroupData(flag, dept)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return
}

// ConvertUserData 将用户信息转成本地结构体
func ConvertUserData(flag string, remoteData []map[string]interface{}) (users []*userModel.User, err error) {
	for _, staff := range remoteData {
		groupIds, err := userModel.DeptIdsToGroupIds(staff["department_ids"].([]string))
		if err != nil {
			return nil, tools.NewMySqlError(fmt.Errorf("将部门ids转换为内部部门id失败：%s", err.Error()))
		}
		user, err := BuildUserData(flag, staff)
		if err != nil {
			return nil, err
		}
		if user != nil {
			user.DepartmentId = tools.SliceToString(groupIds, ",")
			users = append(users, user)
		}
	}
	return
}

func GroupListToTree(rootId string, groupList []*userModel.Group) *userModel.Group {
	// 创建空根节点
	rootGroup := &userModel.Group{SourceDeptId: rootId}
	rootGroup.Children = groupListToTree(rootGroup, groupList)
	return rootGroup
}

func groupListToTree(rootGroup *userModel.Group, list []*userModel.Group) []*userModel.Group {
	children := make([]*userModel.Group, 0)
	for _, group := range list {
		if group.SourceDeptParentId == rootGroup.SourceDeptId {
			children = append(children, group)
		}
	}
	for _, group := range children {
		group.Children = groupListToTree(group, list)
	}
	return children
}

func generateMobile() string {
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(9000000000) + 1000000000
	randNum = randNum + 10000000000
	if userModel.UserSrvIns.Exist(tools.H{"mobile": randNum}) {
		generateMobile()
	}
	return fmt.Sprintf("%v", randNum)
}
