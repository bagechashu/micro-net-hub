package usermgr

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	"github.com/zhaoyunxing92/dingtalk/v2"
	"github.com/zhaoyunxing92/dingtalk/v2/request"

	accountModel "micro-net-hub/internal/module/account/model"
	userProcess "micro-net-hub/internal/module/account/user"
	"micro-net-hub/internal/module/goldap/ldapmgr"
)

type DingTalk struct {
	Client *dingtalk.DingTalk
}

var ding DingTalk

func NewDingTalk() DingTalk {
	once.Do(func() {
		var err error
		ding.Client, err = dingtalk.NewClient(config.Conf.DingTalk.AppKey, config.Conf.DingTalk.AppSecret)
		if err != nil {
			global.Log.Error("init dingding client failed, err:%v\n", err)
		}
	})

	return ding
}

// 通过钉钉获取部门信息
func (mgr DingTalk) SyncDepts() error {
	// 1.获取所有部门
	deptSource, err := mgr.GetAllDepts()
	if err != nil {
		return helper.NewOperationError(fmt.Errorf("获取钉钉部门列表失败：%s", err.Error()))
	}
	depts, err := userProcess.ConvertDeptData(config.Conf.DingTalk.Flag, deptSource)
	if err != nil {
		return helper.NewOperationError(fmt.Errorf("转换钉钉部门数据失败：%s", err.Error()))
	}

	// 2.将远程数据转换成树
	deptTree := userProcess.GroupListToTree(fmt.Sprintf("%s_1", config.Conf.DingTalk.Flag), depts)

	// 3.根据树进行创建
	return mgr.addDeptsRec(deptTree.Children)
}

// 根据现有数据库同步到的部门信息，开启用户同步
func (mgr DingTalk) SyncUsers() error {
	// 1.获取钉钉用户列表
	staffSource, err := mgr.GetAllUsers()
	if err != nil {
		return helper.NewOperationError(fmt.Errorf("SyncDingTalkUsers获取钉钉用户列表失败：%s", err.Error()))
	}
	staffs, err := userProcess.ConvertUserData(config.Conf.DingTalk.Flag, staffSource)
	if err != nil {
		return helper.NewOperationError(fmt.Errorf("转换钉钉用户数据失败：%s", err.Error()))
	}
	// 2.遍历用户，开始写入
	for _, staff := range staffs {
		// 入库
		err = mgr.AddUsers(staff)
		if err != nil {
			return helper.NewOperationError(fmt.Errorf("SyncDingTalkUsers写入用户失败：%s", err.Error()))
		}
	}

	// 3.获取钉钉已离职用户id列表
	// 根据配置判断是查全部离职用户还是只查指定时间范围内的离职用户
	var userIds []string
	if config.Conf.DingTalk.ULeaveRange == 0 {
		userIds, err = mgr.GetLeaveUserIds()
	} else {
		userIds, err = mgr.GetLeaveUserIdsDateRange(config.Conf.DingTalk.ULeaveRange)
	}
	if err != nil {
		return helper.NewOperationError(fmt.Errorf("SyncDingTalkUsers获取钉钉离职用户列表失败：%s", err.Error()))
	}
	// 4.遍历id，开始处理
	for _, uid := range userIds {
		var u accountModel.User
		if u.Exist(
			tools.H{
				"source_user_id": fmt.Sprintf("%s_%s", config.Conf.DingTalk.Flag, uid),
				"status":         1, //只处理1在职的
			}) {
			user := new(accountModel.User)
			err = user.Find(tools.H{"source_user_id": fmt.Sprintf("%s_%s", config.Conf.DingTalk.Flag, uid)})
			if err != nil {
				return helper.NewMySqlError(fmt.Errorf("在MySQL查询用户失败: " + err.Error()))
			}
			// 先从ldap删除用户
			err = ldapmgr.LdapUserDelete(user.UserDN)
			if err != nil {
				return helper.NewLdapError(fmt.Errorf("在LDAP删除用户失败" + err.Error()))
			}
			// 然后更新MySQL中用户状态
			err = user.ChangeStatus(2)
			if err != nil {
				return helper.NewMySqlError(fmt.Errorf("在MySQL更新用户状态失败: " + err.Error()))
			}
		}
	}

	return nil
}

// 官方文档地址： https://open.dingtalk.com/document/orgapp-server/obtain-the-department-list
// GetAllDepts 获取所有部门
func (mgr DingTalk) GetAllDepts() (ret []map[string]interface{}, err error) {
	c := NewDingTalk()
	depts, err := c.Client.FetchDeptList(1, true, "zh_CN")
	if err != nil {
		return ret, err
	}
	if len(config.Conf.DingTalk.DeptList) == 0 {

		ret = make([]map[string]interface{}, 0)
		for _, dept := range depts.Dept {
			ele := make(map[string]interface{})
			ele["id"] = dept.Id
			ele["name"] = dept.Name
			ele["parentid"] = dept.ParentId
			ele["custom_name_pinyin"] = tools.ConvertToPinYin(dept.Name)
			ret = append(ret, ele)
		}
	} else {

		// 遍历配置的部门ID列表获取数据进行处理
		// 从取得的所有部门列表中将配置的部门ID筛选出来再去请求其子部门过滤为1和为配置值的部门ID
		ret = make([]map[string]interface{}, 0)

		for _, dept := range depts.Dept {
			inset := false
			for _, dep_s := range config.Conf.DingTalk.DeptList {
				if strings.HasPrefix(dep_s, "^") {
					continue
				}
				setdepid, _ := strconv.Atoi(dep_s)
				if dept.Id == setdepid {
					inset = true
					break
				}
			}
			if dept.Id == 1 || inset {
				ele := make(map[string]interface{})
				ele["id"] = dept.Id
				ele["name"] = dept.Name
				ele["parentid"] = dept.ParentId
				ele["custom_name_pinyin"] = tools.ConvertToPinYin(dept.Name)
				ret = append(ret, ele)
			}
		}

		for _, dep_s := range config.Conf.DingTalk.DeptList {
			dept_id := dep_s

			if strings.HasPrefix(dep_s, "^") || dept_id == "1" {
				continue
			}
			depid, _ := strconv.Atoi(dept_id)
			depts, err := c.Client.FetchDeptList(depid, true, "zh_CN")

			if err != nil {
				return ret, err
			}

			for _, dept := range depts.Dept {
				ele := make(map[string]interface{})
				ele["id"] = dept.Id
				ele["name"] = dept.Name
				ele["parentid"] = dept.ParentId
				ele["custom_name_pinyin"] = tools.ConvertToPinYin(dept.Name)
				ret = append(ret, ele)
			}
		}
	}
	return
}

// 官方文档地址： https://open.dingtalk.com/document/orgapp-server/queries-the-complete-information-of-a-department-user
// GetAllUsers 获取所有员工信息
func (mgr DingTalk) GetAllUsers() (ret []map[string]interface{}, err error) {
	c := NewDingTalk()
	depts, err := mgr.GetAllDepts()
	if err != nil {
		return nil, err
	}
	for _, dept := range depts {
		r := request.DeptDetailUserInfo{
			DeptId:   dept["id"].(int),
			Cursor:   0,
			Size:     99,
			Language: "zh_CN",
		}
		for {
			//获取钉钉部门人员信息
			rsp, err := c.Client.GetDeptDetailUserInfo(&r)
			if err != nil {
				return nil, err
			}
			for _, user := range rsp.Page.List {
				ele := make(map[string]interface{})
				ele["userid"] = user.UserId
				ele["unionid"] = user.UnionId
				ele["custom_name_pinyin"] = tools.ConvertToPinYin(user.Name)
				ele["name"] = user.Name
				ele["avatar"] = user.Avatar
				ele["mobile"] = user.Mobile
				ele["job_number"] = user.JobNumber
				ele["title"] = user.Title
				ele["work_place"] = user.WorkPlace
				ele["remark"] = user.Remark
				ele["leader"] = user.Leader
				ele["org_email"] = user.OrgEmail
				if user.OrgEmail != "" {
					ele["custom_nickname_org_email"] = strings.Split(user.OrgEmail, "@")[0]
				}
				ele["email"] = user.Email
				if user.Email != "" {
					ele["custom_nickname_email"] = strings.Split(user.Email, "@")[0]
				}
				// 部门ids
				var sourceDeptIds []string
				for _, deptId := range user.DeptIds {
					sourceDeptIds = append(sourceDeptIds, fmt.Sprintf("%s_%d", config.Conf.DingTalk.Flag, deptId))
				}
				ele["department_ids"] = sourceDeptIds
				ret = append(ret, ele)
			}
			if !rsp.Page.HasMore {
				break
			}
			r.Cursor = rsp.Page.NextCursor
		}
	}
	return
}

// 官方文档：https://open.dingtalk.com/document/orgapp-server/intelligent-personnel-query-company-turnover-list
// GetLeaveUserIds 获取离职人员ID列表
func (mgr DingTalk) GetLeaveUserIds() ([]string, error) {
	c := NewDingTalk()
	var ids []string
	ReqParm := struct {
		Cursor int `json:"cursor"`
		Size   int `json:"size"`
	}{
		Cursor: 0,
		Size:   50,
	}

	for {
		rsp, err := c.Client.GetHrmResignEmployeeIds(ReqParm.Cursor, ReqParm.Size)
		if err != nil {
			return nil, err
		}
		ids = append(ids, rsp.Result.UserIds...)
		if rsp.Result.NextCursor == 0 {
			break
		}
		ReqParm.Cursor = rsp.Result.NextCursor
	}
	return ids, nil
}

// 官方文档：https://open.dingtalk.com/document/orgapp/query-the-details-of-employees-who-have-left-office
// GetLeaveUserIdsByDateRange 新接口根据时间范围获取离职人员ID列表
// GetHrmempLeaveRecordsKey    = "/v1.0/contact/empLeaveRecords"
func (mgr DingTalk) GetLeaveUserIdsDateRange(pushDays uint) ([]string, error) {
	c := NewDingTalk()
	var ids []string
	// 配置值为正数,往前推转为负数
	var leaveDays = int(0 - pushDays)
	startTime := time.Now().AddDate(0, 0, leaveDays).Format("2006-01-02T15:04:05Z")
	endTime := time.Now().Format("2006-01-02T15:04:05Z")
	ReqParm := struct {
		StartTime  string `json:"startTime"`
		EndTime    string `json:"endTime"`
		NextToken  string `json:"nextToken"`
		MaxResults int    `json:"maxResults"`
	}{
		StartTime:  startTime,
		EndTime:    endTime,
		NextToken:  "0",
		MaxResults: 50,
	}
	// 使用新的使用时间范围查询离职人员接口获取离职用户ID
	for {
		rsp, err := c.Client.GetHrmEmpLeaveRecords(ReqParm.StartTime, ReqParm.EndTime, ReqParm.NextToken, ReqParm.MaxResults)
		if err != nil {
			return nil, err
		}
		for _, g := range rsp.Records {
			ids = append(ids, g.UserId)
		}

		if rsp.NextToken == "0" || rsp.NextToken == "" {
			break
		}
		ReqParm.NextToken = rsp.NextToken
	}
	return ids, nil
}

// 添加部门
func (mgr DingTalk) addDeptsRec(depts []*accountModel.Group) error {
	for _, dept := range depts {
		err := mgr.AddDept(dept)
		if err != nil {
			return helper.NewOperationError(fmt.Errorf("DsyncDingTalkDepts添加部门失败: %s", err.Error()))
		}
		if len(dept.Children) != 0 {
			err = mgr.addDeptsRec(dept.Children)
			if err != nil {
				return helper.NewOperationError(fmt.Errorf("DsyncDingTalkDepts添加部门失败: %s", err.Error()))
			}
		}
	}
	return nil
}

// AddGroup 添加部门数据
func (mgr DingTalk) AddDept(group *accountModel.Group) error {
	parentGroup := new(accountModel.Group)
	err := parentGroup.Find(tools.H{"source_dept_id": group.SourceDeptParentId}) // 查询当前分组父ID在MySQL中的数据信息
	if err != nil {
		return helper.NewMySqlError(fmt.Errorf("查询父级部门失败：%s", err.Error()))
	}

	// 此时的 group 已经附带了Build后动态关联好的字段，接下来将一些确定性的其他字段值添加上，就可以创建这个分组了
	group.Creator = "system"
	group.GroupType = "cn"
	group.ParentId = parentGroup.ID
	group.Source = config.Conf.DingTalk.Flag
	group.GroupDN = fmt.Sprintf("cn=%s,%s", group.GroupName, parentGroup.GroupDN)

	if !group.Exist(tools.H{"group_dn": group.GroupDN}) { // 判断当前部门是否已落库
		err = userProcess.CommonAddGroup(group)
		if err != nil {
			return helper.NewOperationError(fmt.Errorf("添加部门: %s, 失败: %s", group.GroupName, err.Error()))
		}
	}
	return nil
}

// AddUser 添加用户数据
func (mgr DingTalk) AddUsers(user *accountModel.User) error {
	// 根据角色id获取角色
	roles := accountModel.NewRoles()
	err := roles.GetRolesByIds([]uint{2}) // 默认添加为普通用户角色
	if err != nil {
		return helper.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败:%s", err.Error()))
	}
	user.Roles = roles
	user.Creator = "system"
	user.Source = config.Conf.DingTalk.Flag
	user.Password = config.Conf.Ldap.UserInitPassword
	user.UserDN = fmt.Sprintf("uid=%s,%s", user.Username, config.Conf.Ldap.UserDN)

	// 根据 user_dn 查询用户,不存在则创建
	var gs = accountModel.NewGroups()
	if !user.Exist(tools.H{"user_dn": user.UserDN}) {
		// 获取用户将要添加的分组
		err := gs.GetGroupsByIds(tools.StringToSlice(user.DepartmentIds, ","))
		if err != nil {
			return helper.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败" + err.Error()))
		}
		var deptTmp string
		for _, group := range gs {
			deptTmp = deptTmp + group.GroupName + ","
		}
		user.Departments = strings.TrimRight(deptTmp, ",")

		// 新增用户
		err = userProcess.CommonAddUser(user, gs)
		if err != nil {
			return helper.NewOperationError(fmt.Errorf("添加用户: %s, 失败: %s", user.Username, err.Error()))
		}
	} else {
		if config.Conf.Sync.IsUpdateSyncd {
			// 先获取用户信息
			oldData := new(accountModel.User)
			err = oldData.Find(tools.H{"user_dn": user.UserDN})
			if err != nil {
				return err
			}
			// 获取用户将要添加的分组
			err := gs.GetGroupsByIds(tools.StringToSlice(user.DepartmentIds, ","))
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
			if err = userProcess.CommonUpdateUser(oldData, user, tools.StringToSlice(user.DepartmentIds, ",")); err != nil {
				return err
			}
		}
	}
	return nil
}
