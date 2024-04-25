package group

import (
	"fmt"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"

	"strconv"
	"strings"

	"micro-net-hub/internal/tools"

	"micro-net-hub/internal/module/account/current"
	accountModel "micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/module/goldap/ldapmgr"
)

// List 记录列表
func List(c *gin.Context) {
	req := new(accountModel.GroupListReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.GroupListReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		// 获取数据列表
		groups, err := r.List()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取分组列表失败: %s", err.Error()))
		}

		rets := make([]accountModel.Group, 0)
		for _, group := range groups {
			rets = append(rets, *group)
		}
		count, err := accountModel.GroupCount()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取分组总数失败"))
		}

		return accountModel.GroupListRsp{
			Total:  count,
			Groups: rets,
		}, nil
	})
}

// UserInGroup 在分组内的用户
func UserInGroup(c *gin.Context) {
	req := new(accountModel.UserInGroupReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.UserInGroupReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		filter := tools.H{"id": r.GroupID}
		var g accountModel.Group
		if !g.Exist(filter) {
			return nil, helper.NewMySqlError(fmt.Errorf("分组不存在"))
		}

		group := new(accountModel.Group)
		err := group.Find(filter)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取分组失败: %s", err.Error()))
		}

		rets := make([]accountModel.GuserRsp, 0)

		for _, user := range group.Users {
			if r.Nickname != "" && !strings.Contains(user.Nickname, r.Nickname) {
				continue
			}
			rets = append(rets, accountModel.GuserRsp{
				UserId:       int64(user.ID),
				UserName:     user.Username,
				NickName:     user.Nickname,
				Mail:         user.Mail,
				JobNumber:    user.JobNumber,
				Mobile:       user.Mobile,
				Introduction: user.Introduction,
			})
		}

		return accountModel.GroupUsersRsp{
			GroupId:     int64(group.ID),
			GroupName:   group.GroupName,
			GroupRemark: group.Remark,
			UserList:    rets,
		}, nil
	})
}

// UserNoInGroup 不在分组的用户
func UserNoInGroup(c *gin.Context) {
	req := new(accountModel.UserNoInGroupReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.UserNoInGroupReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		filter := tools.H{"id": r.GroupID}

		var g accountModel.Group
		if !g.Exist(filter) {
			return nil, helper.NewMySqlError(fmt.Errorf("分组不存在"))
		}

		group := new(accountModel.Group)
		err := group.Find(filter)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取分组失败: %s", err.Error()))
		}

		var userList = accountModel.NewUsers()
		err = userList.ListAll()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: " + err.Error()))
		}

		rets := make([]accountModel.GuserRsp, 0)
		for _, user := range userList {
			in := true
			for _, groupUser := range group.Users {
				if user.Username == groupUser.Username {
					in = false
					break
				}
			}
			if in {
				if r.Nickname != "" && !strings.Contains(user.Nickname, r.Nickname) {
					continue
				}
				rets = append(rets, accountModel.GuserRsp{
					UserId:       int64(user.ID),
					UserName:     user.Username,
					NickName:     user.Nickname,
					Mail:         user.Mail,
					JobNumber:    user.JobNumber,
					Mobile:       user.Mobile,
					Introduction: user.Introduction,
				})
			}
		}

		return accountModel.GroupUsersRsp{
			GroupId:     int64(group.ID),
			GroupName:   group.GroupName,
			GroupRemark: group.Remark,
			UserList:    rets,
		}, nil
	})
}

// GetTree 接口树
func GetTree(c *gin.Context) {
	req := new(accountModel.GroupListReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.GroupListReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		var gs = accountModel.NewGroups()
		var err error
		gs, err = r.ListTree()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: " + err.Error()))
		}

		tree := gs.GenGroupTree(0)

		return tree, nil
	})
}

// Add 新建记录
func Add(c *gin.Context) {
	req := new(accountModel.GroupAddReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.GroupAddReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		// 获取当前用户
		ctxUser, err := current.GetCurrentLoginUser(c)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败"))
		}

		group := accountModel.Group{
			GroupType: r.GroupType,
			ParentId:  r.ParentId,
			GroupName: r.GroupName,
			Remark:    r.Remark,
			Creator:   ctxUser.Username,
			Source:    "platform", //默认是平台添加
		}

		if r.ParentId == 0 {
			group.SourceDeptId = "platform_0"
			group.SourceDeptParentId = "platform_0"
			group.GroupDN = fmt.Sprintf("%s=%s,%s", r.GroupType, r.GroupName, config.Conf.Ldap.BaseDN)
		} else {
			parentGroup := new(accountModel.Group)
			err := parentGroup.Find(tools.H{"id": r.ParentId})
			if err != nil {
				return nil, helper.NewMySqlError(fmt.Errorf("获取父级组信息失败"))
			}
			group.SourceDeptId = "platform_0"
			group.SourceDeptParentId = fmt.Sprintf("%s_%d", parentGroup.Source, r.ParentId)
			group.GroupDN = fmt.Sprintf("%s=%s,%s", r.GroupType, r.GroupName, parentGroup.GroupDN)
		}

		// 根据 group_dn 判断分组是否已存在
		var g accountModel.Group
		if g.Exist(tools.H{"group_dn": group.GroupDN}) {
			return nil, helper.NewValidatorError(fmt.Errorf("该分组对应DN已存在"))
		}

		// 先在ldap中创建组
		err = ldapmgr.LdapDeptAdd(&group)
		if err != nil {
			return nil, helper.NewLdapError(fmt.Errorf("向LDAP创建分组失败" + err.Error()))
		}

		// 然后在数据库中创建组
		err = group.Add()
		if err != nil {
			return nil, helper.NewLdapError(fmt.Errorf("向MySQL创建分组失败"))
		}

		// 默认创建分组之后，需要将admin添加到分组中
		adminInfo := new(accountModel.User)
		err = adminInfo.Find(tools.H{"id": 1})
		if err != nil {
			return nil, helper.NewMySqlError(err)
		}

		err = group.AddUserToGroup(adminInfo)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("添加用户到分组失败: %s", err.Error()))
		}

		return nil, nil
	})
}

// Update 更新记录
func Update(c *gin.Context) {
	req := new(accountModel.GroupUpdateReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.GroupUpdateReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		filter := tools.H{"id": int(r.ID)}
		var g accountModel.Group
		if !g.Exist(filter) {
			return nil, helper.NewMySqlError(fmt.Errorf("分组不存在"))
		}

		// 获取当前登陆用户
		ctxUser, err := current.GetCurrentLoginUser(c)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败"))
		}

		oldGroup := new(accountModel.Group)
		err = oldGroup.Find(filter)
		if err != nil {
			return nil, helper.NewMySqlError(err)
		}

		newGroup := accountModel.Group{
			Model:     oldGroup.Model,
			GroupName: r.GroupName,
			Remark:    r.Remark,
			Creator:   ctxUser.Username,
			GroupType: oldGroup.GroupType,
		}

		//若配置了不允许修改分组名称，则不更新分组名称
		if !config.Conf.Ldap.GroupNameModify {
			newGroup.GroupName = oldGroup.GroupName
		}

		err = ldapmgr.LdapDeptUpdate(oldGroup, &newGroup)
		if err != nil {
			return nil, helper.NewLdapError(fmt.Errorf("向LDAP更新分组失败：" + err.Error()))
		}
		err = newGroup.Update()
		if err != nil {
			return nil, helper.NewLdapError(fmt.Errorf("向MySQL更新分组失败"))
		}
		return nil, nil
	})
}

// Delete 删除记录
func Delete(c *gin.Context) {
	req := new(accountModel.GroupDeleteReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.GroupDeleteReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		for _, id := range r.GroupIds {
			filter := tools.H{"id": int(id)}
			var g accountModel.Group
			if !g.Exist(filter) {
				return nil, helper.NewMySqlError(fmt.Errorf("有分组不存在"))
			}
		}

		var gs = accountModel.NewGroups()
		err := gs.GetGroupsByIds(r.GroupIds)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取分组列表失败: %s", err.Error()))
		}

		for _, g := range gs {
			// 判断存在子分组，不允许删除
			filter := tools.H{"parent_id": int(g.ID)}
			if g.Exist(filter) {
				return nil, helper.NewMySqlError(fmt.Errorf("存在子分组，请先删除子分组，再执行该分组的删除操作！"))
			}

			// 删除的时候先从ldap进行删除
			// global.Log.Infof("print groups before delete: %v", g)
			err = ldapmgr.LdapDeptDelete(g.GroupDN)
			if err != nil {
				return nil, helper.NewLdapError(fmt.Errorf("向LDAP删除分组失败：" + err.Error()))
			}
		}

		// 从MySQL中删除
		err = gs.Delete()
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("删除接口失败: %s", err.Error()))
		}

		return nil, nil
	})
}

// AddUser 添加用户
func AddUser(c *gin.Context) {
	req := new(accountModel.GroupAddUserReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.GroupAddUserReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		filter := tools.H{"id": r.GroupID}

		var g accountModel.Group
		if !g.Exist(filter) {
			return nil, helper.NewMySqlError(fmt.Errorf("分组不存在"))
		}

		var users = accountModel.NewUsers()
		err := users.GetUserByIds(r.UserIds)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取用户列表失败: %s", err.Error()))
		}

		group := new(accountModel.Group)
		err = group.Find(filter)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取分组失败: %s", err.Error()))
		}

		if group.GroupDN[:3] == "ou=" {
			return nil, helper.NewMySqlError(fmt.Errorf("ou类型的分组不能添加用户"))
		}

		// 先添加到MySQL
		err = group.AddUsersToGroup(&users)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("添加用户到分组失败: %s", err.Error()))
		}

		// 再往ldap添加
		for _, user := range users {
			err = ldapmgr.LdapDeptAddUserToGroup(group.GroupDN, user.UserDN)
			if err != nil {
				return nil, helper.NewLdapError(fmt.Errorf("向LDAP添加用户到分组失败" + err.Error()))
			}
		}

		for _, user := range users {
			oldData := new(accountModel.User)
			err = oldData.Find(tools.H{"id": user.ID})
			if err != nil {
				return nil, helper.NewMySqlError(err)
			}
			newData := oldData
			// 添加新增的分组ID与部门
			newData.DepartmentId = oldData.DepartmentId + "," + strconv.Itoa(int(r.GroupID))
			newData.Departments = oldData.Departments + "," + group.GroupName
			err = updataUser(newData)
			if err != nil {
				return nil, helper.NewOperationError(fmt.Errorf("处理用户的部门数据失败:" + err.Error()))
			}
		}

		return nil, nil
	})
}

// RemoveUser 移除用户
func RemoveUser(c *gin.Context) {
	req := new(accountModel.GroupRemoveUserReq)
	helper.HandleRequest(c, req, func(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
		r, ok := req.(*accountModel.GroupRemoveUserReq)
		if !ok {
			return nil, helper.ReqAssertErr
		}
		_ = c

		filter := tools.H{"id": r.GroupID}
		var g accountModel.Group
		if !g.Exist(filter) {
			return nil, helper.NewMySqlError(fmt.Errorf("分组不存在"))
		}

		var users = accountModel.NewUsers()
		err := users.GetUserByIds(r.UserIds)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取用户列表失败: %s", err.Error()))
		}

		group := new(accountModel.Group)
		err = group.Find(filter)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("获取分组失败: %s", err.Error()))
		}

		if group.GroupDN[:3] == "ou=" {
			return nil, helper.NewMySqlError(fmt.Errorf("ou类型的分组内没有用户"))
		}

		// 先操作ldap
		for _, user := range users {
			err := ldapmgr.LdapDeptRemoveUserFromGroup(group.GroupDN, user.UserDN)
			if err != nil {
				return nil, helper.NewLdapError(fmt.Errorf("将用户从ldap移除失败" + err.Error()))
			}
		}

		// 再操作MySQL
		err = group.RemoveUsersFromGroup(&users)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("将用户从MySQL移除失败: %s", err.Error()))
		}

		for _, user := range users {
			oldData := new(accountModel.User)
			err = oldData.Find(tools.H{"id": user.ID})
			if err != nil {
				return nil, helper.NewMySqlError(err)
			}
			newData := oldData

			var newDepts []string
			var newDeptIds []string
			// 删掉移除的分组名字
			for _, v := range strings.Split(oldData.Departments, ",") {
				if v != group.GroupName {
					newDepts = append(newDepts, v)
				}
			}
			// 删掉移除的分组id
			for _, v := range strings.Split(oldData.DepartmentId, ",") {
				if v != strconv.Itoa(int(r.GroupID)) {
					newDeptIds = append(newDeptIds, v)
				}
			}

			newData.Departments = strings.Join(newDepts, ",")
			newData.DepartmentId = strings.Join(newDeptIds, ",")
			err = updataUser(newData)
			if err != nil {
				return nil, helper.NewOperationError(fmt.Errorf("处理用户的部门数据失败:" + err.Error()))
			}
		}

		return nil, nil
	})
}

func updataUser(newUser *accountModel.User) error {
	err := newUser.Update()
	if err != nil {
		return helper.NewMySqlError(fmt.Errorf("在MySQL更新用户失败：" + err.Error()))
	}
	return nil
}
