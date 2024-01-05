package usermgr

import (
	"fmt"
	"strings"

	"micro-net-hub/internal/module/goldap/ldapmgr"
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/config"
	"micro-net-hub/internal/tools"

	"github.com/gin-gonic/gin"
	"github.com/wenerme/go-wecom/wecom"
)

type WeChat struct {
	Client *wecom.Client
}

var wechat WeChat

func NewWeChat() WeChat {
	once.Do(func() {
		wechat.Client = wecom.NewClient(wecom.Conf{
			CorpID:     config.Conf.WeCom.CorpID,
			AgentID:    config.Conf.WeCom.AgentID,
			CorpSecret: config.Conf.WeCom.CorpSecret,
		})

	})
	return wechat
}

// 通过企业微信获取部门信息
func (mgr WeChat) SyncDepts(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	// 1.获取所有部门
	deptSource, err := mgr.GetAllDepts()
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("获取企业微信部门列表失败：%s", err.Error()))
	}
	depts, err := userLogic.ConvertDeptData(config.Conf.WeCom.Flag, deptSource)
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("转换企业微信部门数据失败：%s", err.Error()))
	}

	// 2.将远程数据转换成树
	deptTree := userLogic.GroupListToTree(fmt.Sprintf("%s_1", config.Conf.WeCom.Flag), depts)

	// 3.根据树进行创建
	err = mgr.addDeptsRec(deptTree.Children)

	return nil, err
}

// 根据现有数据库同步到的部门信息，开启用户同步
func (mgr WeChat) SyncUsers(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	// 1.获取企业微信用户列表
	staffSource, err := mgr.GetAllUsers()
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("获取企业微信用户列表失败：%s", err.Error()))
	}
	staffs, err := userLogic.ConvertUserData(config.Conf.WeCom.Flag, staffSource)
	if err != nil {
		return nil, tools.NewOperationError(fmt.Errorf("转换企业微信用户数据失败：%s", err.Error()))
	}
	// 2.遍历用户，开始写入
	for _, staff := range staffs {
		// 入库
		err = mgr.AddUsers(staff)
		if err != nil {
			return nil, tools.NewOperationError(fmt.Errorf("SyncWeComUsers写入用户失败：%s", err.Error()))
		}
	}

	// 3.获取企业微信已离职用户id列表
	// 拿到MySQL所有用户数据(来源为 wecom的用户)，远程没有的，则说明被删除了
	// 如果以后企业微信透出了已离职用户列表的接口，则这里可以进行改进
	var res []*userModel.User
	users, err := userModel.UserSrvIns.ListAll()
	if err != nil {
		return nil, tools.NewMySqlError(fmt.Errorf("获取用户列表失败：" + err.Error()))
	}
	for _, user := range users {
		if user.Source != config.Conf.WeCom.Flag {
			continue
		}
		in := true
		for _, staff := range staffs {
			if user.Username == staff.Username {
				in = false
				break
			}
		}
		if in {
			res = append(res, user)
		}
	}
	// 4.遍历id，开始处理
	for _, userTmp := range res {
		user := new(userModel.User)
		err = userModel.UserSrvIns.Find(tools.H{"source_user_id": userTmp.SourceUserId, "status": 1}, user)
		if err != nil {
			return nil, tools.NewMySqlError(fmt.Errorf("在MySQL查询用户失败: " + err.Error()))
		}
		// 先从ldap删除用户
		err = ldapmgr.LdapUserDelete(user.UserDN)
		if err != nil {
			return nil, tools.NewLdapError(fmt.Errorf("在LDAP删除用户失败" + err.Error()))
		}
		// 然后更新MySQL中用户状态
		err = userModel.UserSrvIns.ChangeStatus(int(user.ID), 2)
		if err != nil {
			return nil, tools.NewMySqlError(fmt.Errorf("在MySQL更新用户状态失败: " + err.Error()))
		}
	}
	return nil, nil
}

// 官方文档： https://developer.work.weixin.qq.com/document/path/90208
// GetAllDepts 获取所有部门
func (mgr WeChat) GetAllDepts() (ret []map[string]interface{}, err error) {
	c := NewWeChat()
	depts, err := c.Client.ListDepartment(
		&wecom.ListDepartmentRequest{},
	)
	if err != nil {
		return nil, err
	}
	for _, dept := range depts.Department {
		ele := make(map[string]interface{})
		ele["name"] = dept.Name
		ele["custom_name_pinyin"] = tools.ConvertToPinYin(dept.Name)
		ele["id"] = dept.ID
		ele["name_en"] = dept.NameEn
		ele["parentid"] = dept.ParentID
		ret = append(ret, ele)
	}
	return ret, nil
}

// 官方文档： https://developer.work.weixin.qq.com/document/path/90201
// GetAllUsers 获取所有员工信息
func (mgr WeChat) GetAllUsers() (ret []map[string]interface{}, err error) {
	c := NewWeChat()
	depts, err := mgr.GetAllDepts()
	if err != nil {
		return nil, err
	}
	for _, dept := range depts {
		users, err := c.Client.ListUser(
			&wecom.ListUserRequest{
				DepartmentID: fmt.Sprintf("%d", dept["id"].(int)),
				FetchChild:   "1",
			},
		)
		if err != nil {
			return nil, err
		}
		for _, user := range users.UserList {
			ele := make(map[string]interface{})
			ele["name"] = user.Name
			ele["custom_name_pinyin"] = tools.ConvertToPinYin(user.Name)
			ele["userid"] = user.UserID
			ele["mobile"] = user.Mobile
			ele["position"] = user.Position
			ele["gender"] = user.Gender
			ele["email"] = user.Email
			if user.Email != "" {
				ele["custom_nickname_email"] = strings.Split(user.Email, "@")[0]
			}
			ele["biz_email"] = user.BizMail
			if user.BizMail != "" {
				ele["custom_nickname_biz_email"] = strings.Split(user.BizMail, "@")[0]
			}
			ele["avatar"] = user.Avatar
			ele["telephone"] = user.Telephone
			ele["alias"] = user.Alias
			ele["external_position"] = user.ExternalPosition
			ele["address"] = user.Address
			ele["open_userid"] = user.OpenUserID
			ele["main_department"] = user.MainDepartment
			ele["english_name"] = user.EnglishName
			// 部门ids
			var sourceDeptIds []string
			for _, deptId := range user.Department {
				sourceDeptIds = append(sourceDeptIds, fmt.Sprintf("%s_%d", config.Conf.WeCom.Flag, deptId))
			}
			ele["department_ids"] = sourceDeptIds
			ret = append(ret, ele)
		}
	}
	return ret, nil
}

// 添加部门
func (mgr WeChat) addDeptsRec(depts []*userModel.Group) error {
	for _, dept := range depts {
		err := mgr.AddDepts(dept)
		if err != nil {
			return tools.NewOperationError(fmt.Errorf("DsyncWeComDepts添加部门失败: %s", err.Error()))
		}
		if len(dept.Children) != 0 {
			err = mgr.addDeptsRec(dept.Children)
			if err != nil {
				return tools.NewOperationError(fmt.Errorf("DsyncWeComDepts添加部门失败: %s", err.Error()))
			}
		}
	}
	return nil
}

// AddGroup 添加部门数据
func (mgr WeChat) AddDepts(group *userModel.Group) error {
	// 判断部门名称是否存在
	parentGroup := new(userModel.Group)
	err := parentGroup.Find(tools.H{"source_dept_id": group.SourceDeptParentId})
	if err != nil {
		return tools.NewMySqlError(fmt.Errorf("查询父级部门失败：%s", err.Error()))
	}

	// 此时的 group 已经附带了Build后动态关联好的字段，接下来将一些确定性的其他字段值添加上，就可以创建这个分组了
	group.Creator = "system"
	group.GroupType = "cn"
	group.ParentId = parentGroup.ID
	group.Source = config.Conf.WeCom.Flag
	group.GroupDN = fmt.Sprintf("cn=%s,%s", group.GroupName, parentGroup.GroupDN)

	if !group.Exist(tools.H{"group_dn": group.GroupDN}) {
		err = userLogic.CommonAddGroup(group)
		if err != nil {
			return tools.NewOperationError(fmt.Errorf("添加部门: %s, 失败: %s", group.GroupName, err.Error()))
		}
	}
	return nil
}

// AddUser 添加用户数据
func (mgr WeChat) AddUsers(user *userModel.User) error {
	// 根据角色id获取角色
	roles, err := userModel.RoleSrvIns.GetRolesByIds([]uint{2})
	if err != nil {
		return tools.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败:%s", err.Error()))
	}
	user.Creator = "system"
	user.Roles = roles
	user.Password = config.Conf.Ldap.UserInitPassword
	user.Source = config.Conf.WeCom.Flag
	user.UserDN = fmt.Sprintf("uid=%s,%s", user.Username, config.Conf.Ldap.UserDN)

	// 根据 user_dn 查询用户,不存在则创建
	var gs = userModel.NewGroups()
	if !userModel.UserSrvIns.Exist(tools.H{"user_dn": user.UserDN}) {
		// 获取用户将要添加的分组
		err := gs.GetGroupsByIds(tools.StringToSlice(user.DepartmentId, ","))
		if err != nil {
			return tools.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
		}
		var deptTmp string
		for _, group := range gs {
			deptTmp = deptTmp + group.GroupName + ","
		}
		user.Departments = strings.TrimRight(deptTmp, ",")

		// 创建用户
		err = userLogic.CommonAddUser(user, gs)
		if err != nil {
			return tools.NewOperationError(fmt.Errorf("添加用户: %s, 失败: %s", user.Username, err.Error()))
		}
	} else {
		// 此处逻辑未经实际验证，如在使用中有问题，请反馈
		if config.Conf.WeCom.IsUpdateSyncd {
			// 先获取用户信息
			oldData := new(userModel.User)
			err = userModel.UserSrvIns.Find(tools.H{"user_dn": user.UserDN}, oldData)
			if err != nil {
				return err
			}
			// 获取用户将要添加的分组
			err := gs.GetGroupsByIds(tools.StringToSlice(user.DepartmentId, ","))
			if err != nil {
				return tools.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
			}
			var deptTmp string
			for _, group := range gs {
				deptTmp = deptTmp + group.GroupName + ","
			}
			user.Model = oldData.Model
			user.Roles = oldData.Roles
			user.Creator = oldData.Creator
			user.Source = oldData.Source
			user.Password = oldData.Password
			user.UserDN = oldData.UserDN
			user.Departments = strings.TrimRight(deptTmp, ",")

			// 用户信息的预置处理
			if user.Nickname == "" {
				user.Nickname = oldData.Nickname
			}
			if user.GivenName == "" {
				user.GivenName = user.Nickname
			}
			if user.Introduction == "" {
				user.Introduction = user.Nickname
			}
			if user.Mail == "" {
				user.Mail = oldData.Mail
			}
			if user.JobNumber == "" {
				user.JobNumber = oldData.JobNumber
			}
			if user.Departments == "" {
				user.Departments = oldData.Departments
			}
			if user.Position == "" {
				user.Position = oldData.Position
			}
			if user.PostalAddress == "" {
				user.PostalAddress = oldData.PostalAddress
			}
			if user.Mobile == "" {
				user.Mobile = oldData.Mobile
			}
			if err = userLogic.CommonUpdateUser(oldData, user, tools.StringToSlice(user.DepartmentId, ",")); err != nil {
				return err
			}
		}
	}
	return nil
}
