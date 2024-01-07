package usermgr

import (
	"context"
	"fmt"
	"strings"

	"micro-net-hub/internal/server/config"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	"github.com/chyroc/lark"

	"micro-net-hub/internal/module/goldap/ldapmgr"
	userLogic "micro-net-hub/internal/module/user"
	userModel "micro-net-hub/internal/module/user/model"

	"github.com/gin-gonic/gin"
)

type FeiShu struct {
	Client *lark.Lark
}

var feishu FeiShu

func NewFeiShu() FeiShu {
	once.Do(func() {
		feishu.Client = lark.New(lark.WithAppCredential(
			config.Conf.FeiShu.AppID,
			config.Conf.FeiShu.AppSecret,
		))
	})
	return feishu
}

// 通过飞书获取部门信息
func (mgr FeiShu) SyncDepts(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	// 1.获取所有部门
	deptSource, err := mgr.GetAllDepts()
	if err != nil {
		return nil, helper.NewOperationError(fmt.Errorf("获取飞书部门列表失败：%s", err.Error()))
	}
	depts, err := userLogic.ConvertDeptData(config.Conf.FeiShu.Flag, deptSource)
	if err != nil {
		return nil, helper.NewOperationError(fmt.Errorf("转换飞书部门数据失败：%s", err.Error()))
	}

	// 2.将远程数据转换成树
	deptTree := userLogic.GroupListToTree(fmt.Sprintf("%s_0", config.Conf.FeiShu.Flag), depts)

	// 3.根据树进行创建
	err = mgr.addDeptsRec(deptTree.Children)

	return nil, err
}

// 根据现有数据库同步到的部门信息，开启用户同步
func (mgr FeiShu) SyncUsers(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	// 1.获取飞书用户列表
	staffSource, err := mgr.GetAllUsers()
	if err != nil {
		return nil, helper.NewOperationError(fmt.Errorf("获取飞书用户列表失败：%s", err.Error()))
	}
	staffs, err := userLogic.ConvertUserData(config.Conf.FeiShu.Flag, staffSource)
	if err != nil {
		return nil, helper.NewOperationError(fmt.Errorf("转换飞书用户数据失败：%s", err.Error()))
	}
	// 2.遍历用户，开始写入
	for _, staff := range staffs {
		// 入库
		err = mgr.AddUsers(staff)
		if err != nil {
			return nil, helper.NewOperationError(fmt.Errorf("SyncFeiShuUsers写入用户失败：%s", err.Error()))
		}
	}

	// 3.获取飞书已离职用户id列表
	userIds, err := mgr.GetLeaveUserIds()
	if err != nil {
		return nil, helper.NewOperationError(fmt.Errorf("SyncFeiShuUsers获取飞书离职用户列表失败：%s", err.Error()))
	}
	// 4.遍历id，开始处理
	for _, uid := range userIds {
		var u userModel.User
		if u.Exist(
			tools.H{
				"status":          1, //只处理1在职的
				"source_union_id": fmt.Sprintf("%s_%s", config.Conf.FeiShu.Flag, uid),
			}) {
			user := new(userModel.User)
			err = user.Find(tools.H{"source_union_id": fmt.Sprintf("%s_%s", config.Conf.FeiShu.Flag, uid)})
			if err != nil {
				return nil, helper.NewMySqlError(fmt.Errorf("在MySQL查询用户失败: " + err.Error()))
			}
			// 先从ldap删除用户
			err = ldapmgr.LdapUserDelete(user.UserDN)
			if err != nil {
				return nil, helper.NewLdapError(fmt.Errorf("在LDAP删除用户失败" + err.Error()))
			}
			// 然后更新MySQL中用户状态
			err = user.ChangeStatus(2)
			if err != nil {
				return nil, helper.NewMySqlError(fmt.Errorf("在MySQL更新用户状态失败: " + err.Error()))
			}
		}
	}

	return nil, nil
}

// 官方文档： https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/contact-v3/department/children
// GetAllDepts 获取所有部门
func (mgr FeiShu) GetAllDepts() (ret []map[string]interface{}, err error) {
	c := NewFeiShu()
	var (
		fetchChild bool   = true
		pageSize   int64  = 50
		pageToken  string = ""
		// DeptID     lark.DepartmentIDType = "department_id"
	)

	if len(config.Conf.FeiShu.DeptList) == 0 {
		req := lark.GetDepartmentListReq{
			// DepartmentIDType: &DeptID,
			PageToken:    &pageToken,
			FetchChild:   &fetchChild,
			PageSize:     &pageSize,
			DepartmentID: "0",
		}
		for {
			res, _, err := c.Client.Contact.GetDepartmentList(context.TODO(), &req)
			if err != nil {
				return nil, err
			}

			for _, dept := range res.Items {
				ele := make(map[string]interface{})
				ele["name"] = dept.Name
				ele["custom_name_pinyin"] = tools.ConvertToPinYin(dept.Name)
				ele["parent_department_id"] = dept.ParentDepartmentID
				ele["department_id"] = dept.DepartmentID
				ele["open_department_id"] = dept.OpenDepartmentID
				ele["leader_user_id"] = dept.LeaderUserID
				ele["unit_ids"] = dept.UnitIDs
				ret = append(ret, ele)
			}
			if !res.HasMore {
				break
			}
			pageToken = res.PageToken
		}
	} else {
		//使用dept-list来一个一个添加部门，开头为^的不添加子部门
		isInDeptList := func(id string) bool {
			for _, v := range config.Conf.FeiShu.DeptList {
				if strings.HasPrefix(v, "^") {
					v = v[1:]
				}
				if id == v {
					return true
				}
			}
			return false
		}
		dep_append_norepeat := func(ret []map[string]interface{}, dept map[string]interface{}) []map[string]interface{} {
			for _, v := range ret {
				if v["open_department_id"] == dept["open_department_id"] {
					return ret
				}
			}
			return append(ret, dept)
		}
		for _, dep_s := range config.Conf.FeiShu.DeptList {
			dept_id := dep_s
			no_add_children := false
			if strings.HasPrefix(dep_s, "^") {
				no_add_children = true
				dept_id = dep_s[1:]
			}
			req := lark.GetDepartmentReq{
				DepartmentID: dept_id,
			}
			res, _, err := c.Client.Contact.GetDepartment(context.TODO(), &req)
			if err != nil {
				return nil, err
			}
			ele := make(map[string]interface{})

			ele["name"] = res.Department.Name
			ele["custom_name_pinyin"] = tools.ConvertToPinYin(res.Department.Name)
			if isInDeptList(res.Department.ParentDepartmentID) {
				ele["parent_department_id"] = res.Department.ParentDepartmentID
			} else {
				ele["parent_department_id"] = "0"
			}
			ele["department_id"] = res.Department.DepartmentID
			ele["open_department_id"] = res.Department.OpenDepartmentID
			ele["leader_user_id"] = res.Department.LeaderUserID
			ele["unit_ids"] = res.Department.UnitIDs
			ret = dep_append_norepeat(ret, ele)

			if !no_add_children {
				pageToken = ""
				req := lark.GetDepartmentListReq{
					// DepartmentIDType: &DeptID,
					PageToken:    &pageToken,
					FetchChild:   &fetchChild,
					PageSize:     &pageSize,
					DepartmentID: dept_id,
				}
				for {
					res, _, err := c.Client.Contact.GetDepartmentList(context.TODO(), &req)
					if err != nil {
						return nil, err
					}

					for _, dept := range res.Items {
						ele := make(map[string]interface{})
						ele["name"] = dept.Name
						ele["custom_name_pinyin"] = tools.ConvertToPinYin(dept.Name)
						ele["parent_department_id"] = dept.ParentDepartmentID
						ele["department_id"] = dept.DepartmentID
						ele["open_department_id"] = dept.OpenDepartmentID
						ele["leader_user_id"] = dept.LeaderUserID
						ele["unit_ids"] = dept.UnitIDs
						ret = dep_append_norepeat(ret, ele)
					}
					if !res.HasMore {
						break
					}
					pageToken = res.PageToken
				}
			}
		}
	}
	return
}

// 官方文档： https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/reference/contact-v3/user/find_by_department
// GetAllUsers 获取所有员工信息
func (mgr FeiShu) GetAllUsers() (ret []map[string]interface{}, err error) {
	c := NewFeiShu()
	var (
		pageSize  int64  = 50
		pageToken string = ""
		// deptidtype lark.DepartmentIDType = "department_id"
	)
	depts, err := mgr.GetAllDepts()
	if err != nil {
		return nil, err
	}

	deptids := make([]string, 0)
	// deptids = append(deptids, "0")
	for _, dept := range depts {
		deptids = append(deptids, dept["open_department_id"].(string))
	}

	for _, deptid := range deptids {
		req := lark.GetUserListReq{
			PageSize:  &pageSize,
			PageToken: &pageToken,
			// DepartmentIDType: &deptidtype,
			DepartmentID: deptid,
		}
		for {
			res, _, err := c.Client.Contact.GetUserList(context.Background(), &req)
			if err != nil {
				return nil, err
			}
			for _, user := range res.Items {
				ele := make(map[string]interface{})
				ele["name"] = user.Name
				ele["custom_name_pinyin"] = tools.ConvertToPinYin(user.Name)
				ele["union_id"] = user.UnionID
				ele["user_id"] = user.UserID
				ele["open_id"] = user.OpenID
				ele["en_name"] = user.EnName
				ele["nickname"] = user.Nickname
				if user.Email != "" {
					ele["custom_nickname_email"] = strings.Split(user.Email, "@")[0]
				}
				if user.EnterpriseEmail != "" {
					ele["custom_nickname_enterprise_email"] = strings.Split(user.EnterpriseEmail, "@")[0]
				}
				ele["email"] = user.Email
				ele["mobile"] = user.Mobile
				ele["gender"] = user.Gender
				ele["avatar"] = user.Avatar.AvatarOrigin
				ele["city"] = user.City
				ele["country"] = user.Country
				ele["work_station"] = user.WorkStation
				ele["join_time"] = user.JoinTime
				ele["employee_no"] = user.EmployeeNo
				ele["enterprise_email"] = user.EnterpriseEmail
				ele["job_title"] = user.JobTitle
				// 部门ids
				var sourceDeptIds []string
				for _, deptId := range user.DepartmentIDs {
					sourceDeptIds = append(sourceDeptIds, fmt.Sprintf("%s_%s", config.Conf.FeiShu.Flag, deptId))
				}
				ele["department_ids"] = sourceDeptIds
				ret = append(ret, ele)
			}
			if !res.HasMore {
				pageToken = ""
				break
			}
			pageToken = res.PageToken
		}
	}
	return
}

// 官方文档： https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/ehr/ehr-v1/employee/list
// GetLeaveUserIds 获取离职人员ID列表
func (mgr FeiShu) GetLeaveUserIds() ([]string, error) {
	c := NewFeiShu()
	var ids []string
	users, _, err := c.Client.EHR.GetEHREmployeeList(context.TODO(), &lark.GetEHREmployeeListReq{
		Status:     []int64{5},
		UserIDType: lark.IDTypePtr(lark.IDTypeUnionID), // 只查询unionID
	})
	if err != nil {
		return nil, err
	}
	for _, user := range users.Items {
		ids = append(ids, user.UserID)
	}
	return ids, nil
}

// 添加部门
func (mgr FeiShu) addDeptsRec(depts []*userModel.Group) error {
	for _, dept := range depts {
		err := mgr.AddDepts(dept)
		if err != nil {
			return helper.NewOperationError(fmt.Errorf("DsyncFeiShuDepts添加部门失败: %s", err.Error()))
		}
		if len(dept.Children) != 0 {
			err = mgr.addDeptsRec(dept.Children)
			if err != nil {
				return helper.NewOperationError(fmt.Errorf("DsyncFeiShuDepts添加部门失败: %s", err.Error()))
			}
		}
	}
	return nil
}

// AddGroup 添加部门数据
func (mgr FeiShu) AddDepts(group *userModel.Group) error {
	// 查询当前分组父ID在MySQL中的数据信息
	parentGroup := new(userModel.Group)
	err := parentGroup.Find(tools.H{"source_dept_id": group.SourceDeptParentId})
	if err != nil {
		return helper.NewMySqlError(fmt.Errorf("查询父级部门失败：%s", err.Error()))
	}

	// 此时的 group 已经附带了Build后动态关联好的字段，接下来将一些确定性的其他字段值添加上，就可以创建这个分组了
	group.Creator = "system"
	group.GroupType = "cn"
	group.ParentId = parentGroup.ID
	group.Source = config.Conf.FeiShu.Flag
	group.GroupDN = fmt.Sprintf("cn=%s,%s", group.GroupName, parentGroup.GroupDN)

	if !group.Exist(tools.H{"group_dn": group.GroupDN}) {
		err = userLogic.CommonAddGroup(group)
		if err != nil {
			return helper.NewOperationError(fmt.Errorf("添加部门: %s, 失败: %s", group.GroupName, err.Error()))
		}
	}
	return nil
}

// AddUser 添加用户数据
func (mgr FeiShu) AddUsers(user *userModel.User) error {
	// 根据角色id获取角色
	roles := userModel.NewRoles()
	err := roles.GetRolesByIds([]uint{2})
	if err != nil {
		return helper.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败:%s", err.Error()))
	}
	user.Roles = roles
	user.Creator = "system"
	user.Source = config.Conf.FeiShu.Flag
	user.Password = config.Conf.Ldap.UserInitPassword
	user.UserDN = fmt.Sprintf("uid=%s,%s", user.Username, config.Conf.Ldap.UserDN)

	// 根据 user_dn 查询用户,不存在则创建
	var gs = userModel.NewGroups()
	if !user.Exist(tools.H{"user_dn": user.UserDN}) {
		// 获取用户将要添加的分组
		err := gs.GetGroupsByIds(tools.StringToSlice(user.DepartmentId, ","))
		if err != nil {
			return helper.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
		}
		var deptTmp string
		for _, group := range gs {
			deptTmp = deptTmp + group.GroupName + ","
		}
		user.Departments = strings.TrimRight(deptTmp, ",")

		// 添加用户
		err = userLogic.CommonAddUser(user, gs)
		if err != nil {
			return helper.NewOperationError(fmt.Errorf("添加用户: %s, 失败: %s", user.Username, err.Error()))
		}
	} else {
		// 此处逻辑未经实际验证，如在使用中有问题，请反馈
		if config.Conf.FeiShu.IsUpdateSyncd {
			// 先获取用户信息
			oldData := new(userModel.User)
			err = oldData.Find(tools.H{"user_dn": user.UserDN})
			if err != nil {
				return err
			}
			// 获取用户将要添加的分组
			err := gs.GetGroupsByIds(tools.StringToSlice(user.DepartmentId, ","))
			if err != nil {
				return helper.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
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
