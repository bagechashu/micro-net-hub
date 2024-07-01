package group

import (
	"fmt"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/server/helper"

	"github.com/gin-gonic/gin"

	"strings"

	"micro-net-hub/internal/module/account/auth"
	"micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/module/goldap/ldapmgr"
)

// GroupListReq 获取资源列表结构体
type GroupListReq struct {
	GroupName string                `json:"groupName" form:"groupName"`
	Remark    string                `json:"remark" form:"remark"`
	PageNum   int                   `json:"pageNum" form:"pageNum"`
	PageSize  int                   `json:"pageSize" form:"pageSize"`
	SyncState model.GroupSyncStatus `json:"syncState" form:"syncState"`
}

type GroupListRsp struct {
	Total  int64         `json:"total"`
	Groups []model.Group `json:"groups"`
}

// List 记录列表
func List(c *gin.Context) {
	var req GroupListReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取数据列表
	gs := model.NewGroups()
	err = gs.List(
		&model.Group{
			GroupName: req.GroupName,
			Remark:    req.Remark,
			SyncState: req.SyncState,
		},
		req.PageNum,
		req.PageSize,
	)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取分组列表失败: %s", err.Error())))
		return
	}

	rets := make([]model.Group, 0)
	for _, g := range gs {
		rets = append(rets, *g)
	}
	count, err := model.GroupCount()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取分组总数失败")))
		return
	}

	helper.Success(c, GroupListRsp{
		Total:  count,
		Groups: rets,
	})

}

// GroupListReq 获取资源列表结构体
type GroupListTreeReq struct {
	GroupName string `json:"groupName" form:"groupName"`
	Remark    string `json:"remark" form:"remark"`
	PageNum   int    `json:"pageNum" form:"pageNum"`
	PageSize  int    `json:"pageSize" form:"pageSize"`
}

// GetTree 接口树
func GetTree(c *gin.Context) {
	var req GroupListTreeReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取数据列表
	gs := model.NewGroups()
	err = gs.ListTree(
		&model.Group{
			GroupName: req.GroupName,
			Remark:    req.Remark,
		},
		req.PageNum,
		req.PageSize,
	)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: "+err.Error())))
		return
	}

	tree := gs.GenGroupTree(0)

	helper.Success(c, tree)
}

// UserInGroupReq 在分组内的用户
type UserInGroupReq struct {
	GroupID  uint   `json:"groupId" form:"groupId" validate:"required"`
	Nickname string `json:"nickname" form:"nickname"`
}

type GuserRsp struct {
	UserId       int64  `json:"userId"`
	UserName     string `json:"userName"`
	NickName     string `json:"nickName"`
	Mail         string `json:"mail"`
	JobNumber    string `json:"jobNumber"`
	Mobile       string `json:"mobile"`
	Introduction string `json:"introduction"`
}

type GroupUsersRsp struct {
	GroupId     int64      `json:"groupId"`
	GroupName   string     `json:"groupName"`
	GroupRemark string     `json:"groupRemark"`
	UserList    []GuserRsp `json:"userList"`
}

// UserInGroup 在分组内的用户
func UserInGroup(c *gin.Context) {
	var req UserInGroupReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	filter := map[string]interface{}{"id": req.GroupID}
	var g model.Group
	if !g.Exist(filter) {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("分组不存在")))
		return
	}

	group := new(model.Group)
	err = group.Find(filter)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取分组失败: %s", err.Error())))
		return
	}

	rets := make([]GuserRsp, 0)

	for _, user := range group.Users {
		if req.Nickname != "" && !strings.Contains(user.Nickname, req.Nickname) {
			continue
		}
		rets = append(rets, GuserRsp{
			UserId:       int64(user.ID),
			UserName:     user.Username,
			NickName:     user.Nickname,
			Mail:         user.Mail,
			JobNumber:    user.JobNumber,
			Mobile:       user.Mobile,
			Introduction: user.Introduction,
		})
	}

	helper.Success(c, GroupUsersRsp{
		GroupId:     int64(group.ID),
		GroupName:   group.GroupName,
		GroupRemark: group.Remark,
		UserList:    rets,
	})
}

// UserNoInGroupReq 不在分组内的用户
type UserNoInGroupReq struct {
	GroupID  uint   `json:"groupId" form:"groupId" validate:"required"`
	Nickname string `json:"nickname" form:"nickname"`
}

// UserNoInGroup 不在分组的用户
func UserNoInGroup(c *gin.Context) {
	var req UserNoInGroupReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	filter := map[string]interface{}{"id": req.GroupID}

	var g model.Group
	if !g.Exist(filter) {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("分组不存在")))
		return
	}

	group := new(model.Group)
	err = group.Find(filter)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取分组失败: %s", err.Error())))
		return
	}

	var userList = model.NewUsers()
	err = userList.ListAll()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取资源列表失败: "+err.Error())))
		return
	}

	rets := make([]GuserRsp, 0)
	for _, user := range userList {
		in := true
		for _, groupUser := range group.Users {
			if user.Username == groupUser.Username {
				in = false
				break
			}
		}
		if in {
			if req.Nickname != "" && !strings.Contains(user.Nickname, req.Nickname) {
				continue
			}
			rets = append(rets, GuserRsp{
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

	helper.Success(c, GroupUsersRsp{
		GroupId:     int64(group.ID),
		GroupName:   group.GroupName,
		GroupRemark: group.Remark,
		UserList:    rets,
	})
}

// GroupAddReq 添加资源结构体
type GroupAddReq struct {
	GroupType string `json:"groupType" validate:"required,min=1,max=20"`
	GroupName string `json:"groupName" validate:"required,min=1,max=128"`
	//父级Id 大于等于0 必填
	ParentId uint   `json:"parentId" validate:"omitempty,min=0"`
	Remark   string `json:"remark" validate:"min=0,max=128"` // 分组的中文描述
}

// Add 新建记录
func Add(c *gin.Context) {
	var req GroupAddReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 获取当前用户
	ctxUser, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户信息失败")))
		return
	}

	group := model.Group{
		GroupType: req.GroupType,
		ParentId:  req.ParentId,
		GroupName: req.GroupName,
		Remark:    req.Remark,
		Creator:   ctxUser.Username,
		Source:    "platform", //默认是平台添加
	}

	if req.ParentId == 0 {
		group.SourceDeptId = "platform_0"
		group.SourceDeptParentId = "platform_0"
		group.GroupDN = fmt.Sprintf("%s=%s,%s", req.GroupType, req.GroupName, config.Conf.Ldap.BaseDN)
	} else {
		parentGroup := new(model.Group)
		err := parentGroup.Find(map[string]interface{}{"id": req.ParentId})
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取父级组信息失败")))
			return
		}
		group.SourceDeptId = "platform_0"
		group.SourceDeptParentId = fmt.Sprintf("%s_%d", parentGroup.Source, req.ParentId)
		group.GroupDN = fmt.Sprintf("%s=%s,%s", req.GroupType, req.GroupName, parentGroup.GroupDN)
	}

	// 根据 group_dn 判断分组是否已存在
	var g model.Group
	if g.Exist(map[string]interface{}{"group_dn": group.GroupDN}) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("该分组对应DN已存在")))
		return
	}

	// 先在ldap中创建组
	err = ldapmgr.LdapDeptAdd(&group)
	if err != nil {
		helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("向LDAP创建分组失败"+err.Error())))
		return
	}

	// 然后在数据库中创建组
	err = group.Add()
	if err != nil {
		helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("向MySQL创建分组失败")))
		return
	}

	// 默认创建分组之后，需要将admin添加到分组中
	adminInfo := new(model.User)
	err = adminInfo.Find(map[string]interface{}{"id": 1})
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(err))
		return
	}

	err = group.AddUserToGroup(adminInfo)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("添加用户到分组失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

// GroupUpdateReq 更新资源结构体
type GroupUpdateReq struct {
	ID        uint   `json:"id" form:"id" validate:"required"`
	GroupName string `json:"groupName" validate:"required,min=1,max=128"`
	Remark    string `json:"remark" validate:"min=0,max=128"` // 分组的中文描述
}

// Update 更新记录
func Update(c *gin.Context) {
	var req GroupUpdateReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	filter := map[string]interface{}{"id": int(req.ID)}
	var g model.Group
	if !g.Exist(filter) {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("分组不存在")))
		return
	}

	// 获取当前登陆用户
	ctxUser, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败")))
		return
	}

	oldGroup := new(model.Group)
	err = oldGroup.Find(filter)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(err))
		return
	}

	newGroup := model.Group{
		Model:     oldGroup.Model,
		GroupName: req.GroupName,
		Remark:    req.Remark,
		Creator:   ctxUser.Username,
		GroupType: oldGroup.GroupType,
	}

	//若配置了不允许修改分组名称，则不更新分组名称
	if !config.Conf.Ldap.GroupNameModify {
		newGroup.GroupName = oldGroup.GroupName
	}

	err = ldapmgr.LdapDeptUpdate(oldGroup, &newGroup)
	if err != nil {
		helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("向LDAP更新分组失败："+err.Error())))
		return
	}
	err = newGroup.Update()
	if err != nil {
		helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("向MySQL更新分组失败")))
		return
	}
	helper.Success(c, nil)
}

// GroupDeleteReq 删除资源结构体
type GroupDeleteReq struct {
	GroupIds []uint `json:"groupIds" validate:"required"`
}

// Delete 删除记录
func Delete(c *gin.Context) {
	var req GroupDeleteReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	for _, id := range req.GroupIds {
		filter := map[string]interface{}{"id": int(id)}
		var g model.Group
		if !g.Exist(filter) {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("有分组不存在")))
			return
		}
	}

	var gs = model.NewGroups()
	err = gs.GetGroupsByIds(req.GroupIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取分组列表失败: %s", err.Error())))
		return
	}

	for _, g := range gs {
		// 判断存在子分组，不允许删除
		filter := map[string]interface{}{"parent_id": int(g.ID)}
		if g.Exist(filter) {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("存在子分组，请先删除子分组，再执行该分组的删除操作！")))
			return
		}

		// 删除的时候先从ldap进行删除
		// global.Log.Infof("print groups before delete: %v", g)
		err = ldapmgr.LdapDeptDelete(g.GroupDN)
		if err != nil {
			helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("向LDAP删除分组失败："+err.Error())))
			return
		}
	}

	// 从MySQL中删除
	err = gs.Delete()
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("删除接口失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}

type GroupAddUserReq struct {
	GroupID uint   `json:"groupId" validate:"required"`
	UserIds []uint `json:"userIds" validate:"required"`
}

// AddUser 添加用户
func AddUser(c *gin.Context) {
	var req GroupAddUserReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	filter := map[string]interface{}{"id": req.GroupID}

	var g model.Group
	if !g.Exist(filter) {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("分组不存在")))
		return
	}

	var users = model.NewUsers()
	err = users.GetUserByIds(req.UserIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取用户列表失败: %s", err.Error())))
		return
	}

	group := new(model.Group)
	err = group.Find(filter)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取分组失败: %s", err.Error())))
		return
	}

	if group.GroupDN[:3] == "ou=" {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("ou类型的分组不能添加用户")))
		return
	}

	// 先添加到MySQL
	err = group.AddUsersToGroup(&users)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("添加用户到分组失败: %s", err.Error())))
		return
	}

	// 再往ldap添加
	for _, user := range users {
		err = ldapmgr.LdapDeptAddUserToGroup(group.GroupDN, user.UserDN)
		if err != nil {
			helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("向LDAP添加用户到分组失败"+err.Error())))
			return
		}
	}

	helper.Success(c, nil)
}

type GroupRemoveUserReq struct {
	GroupID uint   `json:"groupId" validate:"required"`
	UserIds []uint `json:"userIds" validate:"required"`
}

// RemoveUser 移除用户
func RemoveUser(c *gin.Context) {
	var req GroupRemoveUserReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	filter := map[string]interface{}{"id": req.GroupID}
	var g model.Group
	if !g.Exist(filter) {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("分组不存在")))
		return
	}

	var users = model.NewUsers()
	err = users.GetUserByIds(req.UserIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取用户列表失败: %s", err.Error())))
		return
	}

	group := new(model.Group)
	err = group.Find(filter)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取分组失败: %s", err.Error())))
		return
	}
	// 检查 GroupDN 是否为 ou 类型
	if group.GroupDN[:3] == "ou=" {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("ou类型的分组内没有用户")))
		return
	}

	// 先操作ldap
	for _, user := range users {
		err := ldapmgr.LdapDeptRemoveUserFromGroup(group.GroupDN, user.UserDN)
		if err != nil {
			helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("将用户从ldap移除失败"+err.Error())))
			return
		}
	}

	// 再操作MySQL
	err = group.RemoveUsersFromGroup(&users)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("将用户从MySQL移除失败: %s", err.Error())))
		return
	}

	helper.Success(c, nil)
}
