package user

import (
	"fmt"
	"math/rand"
	"time"

	"micro-net-hub/internal/global"
	accountModel "micro-net-hub/internal/module/account/model"
	fieldRelationModel "micro-net-hub/internal/module/goldap/field_relation/model"
	"micro-net-hub/internal/module/goldap/ldapmgr"
	"micro-net-hub/internal/tools"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/server/helper"

	"github.com/tidwall/gjson"
)

// CommonAddGroup 标准创建分组
func CommonAddGroup(group *accountModel.Group) error {
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
	adminInfo := new(accountModel.User)
	err = adminInfo.Find(map[string]interface{}{"id": 1})
	if err != nil {
		return err
	}

	err = group.AddUserToGroup(adminInfo)
	if err != nil {
		return err
	}

	return nil
}

// CommonUpdateGroup 标准更新分组
func CommonUpdateGroup(oldGroup, newGroup *accountModel.Group) error {
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
func CommonAddUser(user *accountModel.User) error {
	user.CheckAttrVacancies()
	if complex := tools.CheckPasswordComplexity(user.Password); !complex {
		return tools.ErrPasswordNotComplex
	}

	// 先将用户添加到MySQL
	err := user.Add()
	if err != nil {
		return helper.NewMySqlError(fmt.Errorf("向MySQL创建用户失败：" + err.Error()))
	}
	// 再将用户添加到ldap
	err = ldapmgr.LdapUserAdd(user)
	if err != nil {
		return helper.NewLdapError(fmt.Errorf("AddUser向LDAP创建用户失败：" + err.Error()))
	}

	// 处理用户归属的组
	for _, group := range user.Groups {
		if group.GroupDN[:3] == "ou=" {
			continue
		}
		//根据选择的部门，添加到部门内
		err = ldapmgr.LdapDeptAddUserToGroup(group.GroupDN, user.UserDN)
		if err != nil {
			return helper.NewLdapError(fmt.Errorf("向Ldap添加用户到分组关系失败：" + err.Error()))
		}
	}
	return nil
}

// CommonUpdateUser 标准更新用户
func CommonUpdateUser(oldUser, newUser *accountModel.User, groupIds []uint) error {
	// 更新用户
	if !config.Conf.Ldap.UserNameModify {
		newUser.Username = oldUser.Username
	}

	newUser.CheckAttrVacancies()

	if newUser.Password != "" {
		if complex := tools.CheckPasswordComplexity(newUser.Password); !complex {
			return tools.ErrPasswordNotComplex
		}
	}

	err := ldapmgr.LdapUserUpdate(oldUser.Username, newUser)
	if err != nil {
		return helper.NewLdapError(fmt.Errorf("在LDAP更新用户失败：" + err.Error()))
	}

	err = newUser.Update()
	if err != nil {
		return helper.NewMySqlError(fmt.Errorf("在MySQL更新用户失败：" + err.Error()))
	}

	//判断部门信息是否有变化有变化则更新相应的数据库
	var oldDeptIds []uint
	for _, group := range oldUser.Groups {
		oldDeptIds = append(oldDeptIds, group.ID)
	}
	addDeptIds, removeDeptIds := tools.ArrUintCmp(oldDeptIds, groupIds)

	// 先处理添加的部门
	var addGroups = accountModel.NewGroups()
	err = addGroups.GetGroupsByIds(addDeptIds)
	if err != nil {
		return helper.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
	}
	for _, group := range addGroups {
		if group.GroupDN[:3] == "ou=" {
			continue
		}
		// 先将用户和部门信息维护到MySQL
		err := group.AddUserToGroup(newUser)
		if err != nil {
			return helper.NewMySqlError(fmt.Errorf("向MySQL添加用户到分组关系失败：" + err.Error()))
		}
		//根据选择的部门，添加到部门内
		err = ldapmgr.LdapDeptAddUserToGroup(group.GroupDN, newUser.UserDN)
		if err != nil {
			return helper.NewLdapError(fmt.Errorf("向Ldap添加用户到分组关系失败：" + err.Error()))
		}
	}

	// 再处理删除的部门
	var removeGroups = accountModel.NewGroups()
	err = removeGroups.GetGroupsByIds(removeDeptIds)
	if err != nil {
		return helper.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
	}
	for _, group := range removeGroups {
		if group.GroupDN[:3] == "ou=" {
			continue
		}
		err := group.RemoveUserFromGroup(newUser)
		if err != nil {
			return helper.NewMySqlError(fmt.Errorf("在MySQL将用户从分组移除失败：" + err.Error()))
		}
		err = ldapmgr.LdapDeptRemoveUserFromGroup(group.GroupDN, newUser.UserDN)
		if err != nil {
			return helper.NewMySqlError(fmt.Errorf("在ldap将用户从分组移除失败：" + err.Error()))
		}
	}
	return nil
}

// ConvertDeptData 将部门信息转成本地结构体
func ConvertDeptData(flag string, remoteData []map[string]interface{}) (groups []*accountModel.Group, err error) {
	for _, dept := range remoteData {
		group, err := buildGroupData(flag, dept)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return
}

// buildGroupData 根据数据与动态字段组装成分组数据
func buildGroupData(flag string, remoteData map[string]interface{}) (*accountModel.Group, error) {
	output, err := json.Marshal(&remoteData)
	if err != nil {
		return nil, err
	}

	oldData := new(fieldRelationModel.FieldRelation)
	err = fieldRelationModel.Find(map[string]interface{}{"flag": flag + "_group"}, oldData)
	if err != nil {
		return nil, helper.NewMySqlError(err)
	}
	frs, err := tools.JsonToMap(string(oldData.Attributes))
	if err != nil {
		return nil, helper.NewOperationError(err)
	}

	g := &accountModel.Group{}
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

// ConvertUserData 将用户信息转成本地结构体
func ConvertUserData(flag string, remoteData []map[string]interface{}) (users []*accountModel.User, err error) {
	for _, staff := range remoteData {
		gs, err := accountModel.ThirdPartDeptIdsToGroups(staff["department_ids"].([]string))
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("将部门ids转换为内部部门id失败：%s", err.Error()))
		}
		user, err := buildUserData(flag, staff)
		if err != nil {
			return nil, err
		}
		if user != nil {
			user.Groups = gs
			users = append(users, user)
		}
	}
	return
}

// buildUserData 根据数据与动态字段组装成用户数据
func buildUserData(flag string, remoteData map[string]interface{}) (*accountModel.User, error) {
	output, err := json.Marshal(&remoteData)
	if err != nil {
		return nil, err
	}

	fieldRelationSource := new(fieldRelationModel.FieldRelation)
	err = fieldRelationModel.Find(map[string]interface{}{"flag": flag + "_user"}, fieldRelationSource)
	if err != nil {
		return nil, helper.NewMySqlError(err)
	}
	fieldRelation, err := tools.JsonToMap(string(fieldRelationSource.Attributes))
	if err != nil {
		return nil, helper.NewOperationError(err)
	}

	// 校验username是否为空，username为必填项
	name := gjson.Get(string(output), fieldRelation["username"]).String()
	if len(name) == 0 {
		global.Log.Warnf("%s 该用户未填写username", output)
		return nil, nil
	}

	u := &accountModel.User{}
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
func GroupListToTree(rootId string, groupList []*accountModel.Group) *accountModel.Group {
	// 创建空根节点
	rootGroup := &accountModel.Group{SourceDeptId: rootId}
	rootGroup.Children = groupListToTree(rootGroup, groupList)
	return rootGroup
}

func groupListToTree(rootGroup *accountModel.Group, list []*accountModel.Group) []*accountModel.Group {
	children := make([]*accountModel.Group, 0)
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
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randNum := r.Intn(99999999) + 10000000000
	var u accountModel.User
	if u.Exist(map[string]interface{}{"mobile": randNum}) {
		generateMobile()
	}
	return fmt.Sprintf("%v", randNum)
}
