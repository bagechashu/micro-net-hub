package model

import (
	"errors"
	"fmt"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/tools"
	"strings"

	"gorm.io/gorm"
)

type Group struct {
	gorm.Model
	GroupName          string   `gorm:"type:varchar(128);comment:'分组名称'" json:"groupName"`
	Remark             string   `gorm:"type:varchar(128);comment:'分组中文说明'" json:"remark"`
	Creator            string   `gorm:"type:varchar(20);comment:'创建人'" json:"creator"`
	GroupType          string   `gorm:"type:varchar(20);comment:'分组类型：cn、ou'" json:"groupType"`
	Users              []*User  `gorm:"many2many:group_users" json:"users"`
	ParentId           uint     `gorm:"default:0;comment:'父组编号(编号为0时表示根组)'" json:"parentId"`
	SourceDeptId       string   `gorm:"type:varchar(100);comment:'部门编号'" json:"sourceDeptId"`
	Source             string   `gorm:"type:varchar(20);comment:'来源：dingTalk、weCom、ldap、platform'" json:"source"`
	SourceDeptParentId string   `gorm:"type:varchar(100);comment:'父部门编号'" json:"sourceDeptParentId"`
	SourceUserNum      int      `gorm:"default:0;comment:'部门下的用户数量，从第三方获取的数据'" json:"source_user_num"`
	Children           []*Group `gorm:"-" json:"children"`
	GroupDN            string   `gorm:"type:varchar(255);not null;comment:'分组dn'" json:"groupDn"`             // 分组在ldap的dn
	SyncState          uint     `gorm:"type:tinyint(1);default:1;comment:'同步状态:1已同步, 2未同步'" json:"syncState"` // 数据到ldap的同步状态
}

func (g *Group) SetGroupName(groupName string) {
	g.GroupName = groupName
}

func (g *Group) SetRemark(remark string) {
	g.Remark = remark
}

func (g *Group) SetSourceDeptId(sourceDeptId string) {
	g.SourceDeptId = sourceDeptId
}

func (g *Group) SetSourceDeptParentId(sourceDeptParentId string) {
	g.SourceDeptParentId = sourceDeptParentId
}

// Add 添加资源
func (g *Group) Add() error {
	return global.DB.Create(g).Error
}

// Update 更新资源
func (g *Group) Update() error {
	return global.DB.Model(g).Where("id = ?", g.ID).Updates(g).Error
}

// ChangeSyncState 更新分组的同步状态
func (g *Group) ChangeSyncState(status int) error {
	return global.DB.Model(&Group{}).Where("id = ?", g.ID).Update("sync_state", status).Error
}

// Find 获取单个资源
func (g *Group) Find(filter map[string]interface{}, args ...interface{}) error {
	return global.DB.Where(filter, args).Preload("Users").First(&g).Error
}

// Exist 判断资源是否存在
func (g *Group) Exist(filter map[string]interface{}) bool {
	err := global.DB.Debug().Order("created_at DESC").Where(filter).First(&g).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// AddUserToGroup 添加用户到分组
func (g *Group) AddUserToGroup(users []User) error {
	return global.DB.Model(&g).Association("Users").Append(users)
}

// RemoveUserFromGroup 将用户从分组移除
func (g *Group) RemoveUserFromGroup(users []User) error {
	return global.DB.Model(&g).Association("Users").Delete(users)
}

// GroupCount 获取数据总数
func GroupCount() (int64, error) {
	var count int64
	err := global.DB.Model(&Group{}).Count(&count).Error
	return count, err
}

// DeptIdsToGroupIds 将企业IM部门id转换为MySQL分组id
func DeptIdsToGroupIds(ids []string) (groupIds []uint, err error) {
	var tempGroups []Group
	err = global.DB.Model(&Group{}).Where("source_dept_id IN (?)", ids).Find(&tempGroups).Error
	if err != nil {
		return nil, err
	}
	var tempGroupIds []uint
	for _, g := range tempGroups {
		tempGroupIds = append(tempGroupIds, g.ID)
	}
	return tempGroupIds, nil
}

type Groups []*Group

func NewGroups() Groups {
	return make([]*Group, 0)
}

// List 获取数据列表
func (gs *Groups) ListAll() (err error) {
	err = global.DB.Model(&Group{}).Order("created_at DESC").Find(&gs).Error
	return err
}

// GenGroupTree 生成分组树
func (gs *Groups) GenGroupTree(parentId uint) []*Group {
	tree := make([]*Group, 0)

	for _, g := range *gs {
		if g.ParentId == parentId {
			children := gs.GenGroupTree(g.ID)
			g.Children = children
			tree = append(tree, g)
		}
	}
	return tree
}

// Delete 批量删除
func (gs *Groups) Delete() error {
	for _, g := range *gs {
		if err := global.DB.Debug().Unscoped().Delete(&g).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetApisById 根据接口ID获取接口列表
func (gs *Groups) GetGroupsByIds(ids []uint) (err error) {
	err = global.DB.Where("id IN (?)", ids).Find(&gs).Error
	return err
}

// GroupListReq 获取资源列表结构体
type GroupListReq struct {
	GroupName string `json:"groupName" form:"groupName"`
	Remark    string `json:"remark" form:"remark"`
	PageNum   int    `json:"pageNum" form:"pageNum"`
	PageSize  int    `json:"pageSize" form:"pageSize"`
	SyncState uint   `json:"syncState" form:"syncState"`
}

// List 获取数据列表
func (req GroupListReq) List() ([]*Group, error) {
	var list []*Group
	db := global.DB.Model(&Group{}).Order("created_at DESC")

	groupName := strings.TrimSpace(req.GroupName)
	if groupName != "" {
		db = db.Where("group_name LIKE ?", fmt.Sprintf("%%%s%%", groupName))
	}
	groupRemark := strings.TrimSpace(req.Remark)
	if groupRemark != "" {
		db = db.Where("remark LIKE ?", fmt.Sprintf("%%%s%%", groupRemark))
	}
	syncState := req.SyncState
	if syncState != 0 {
		db = db.Where("sync_state = ?", syncState)
	}

	pageReq := tools.NewPageOption(req.PageNum, req.PageSize)
	err := db.Offset(pageReq.PageNum).Limit(pageReq.PageSize).Preload("Users").Find(&list).Error
	return list, err
}

// List 获取数据列表
func (req GroupListReq) ListTree() ([]*Group, error) {
	var gs = NewGroups()
	db := global.DB.Model(&Group{}).Order("created_at DESC")

	groupName := strings.TrimSpace(req.GroupName)
	if groupName != "" {
		db = db.Where("group_name LIKE ?", fmt.Sprintf("%%%s%%", groupName))
	}
	groupRemark := strings.TrimSpace(req.Remark)
	if groupRemark != "" {
		db = db.Where("remark LIKE ?", fmt.Sprintf("%%%s%%", groupRemark))
	}

	pageReq := tools.NewPageOption(req.PageNum, req.PageSize)
	err := db.Offset(pageReq.PageNum).Limit(pageReq.PageSize).Find(&gs).Error
	return gs, err
}

// GroupListAllReq 获取资源列表结构体，不分页
type GroupListAllReq struct {
	GroupName          string `json:"groupName" form:"groupName"`
	GroupType          string `json:"groupType" form:"groupType"`
	Remark             string `json:"remark" form:"remark"`
	Source             string `json:"source" form:"source"`
	SourceDeptId       string `json:"sourceDeptId"`
	SourceDeptParentId string `json:"SourceDeptParentId"`
}

// GroupAddReq 添加资源结构体
type GroupAddReq struct {
	GroupType string `json:"groupType" validate:"required,min=1,max=20"`
	GroupName string `json:"groupName" validate:"required,min=1,max=128"`
	//父级Id 大于等于0 必填
	ParentId uint   `json:"parentId" validate:"omitempty,min=0"`
	Remark   string `json:"remark" validate:"min=0,max=128"` // 分组的中文描述
}

// DingTalkGroupAddReq 添加钉钉资源结构体
type DingGroupAddReq struct {
	GroupType string `json:"groupType" validate:"required,min=1,max=20"`
	GroupName string `json:"groupName" validate:"required,min=1,max=128"`
	//父级Id 大于等于0 必填
	ParentId           uint   `json:"parentId" validate:"omitempty,min=0"`
	Remark             string `json:"remark" validate:"min=0,max=128"` // 分组的中文描述
	SourceDeptId       string `json:"sourceDeptId"`
	Source             string `json:"source"`
	SourceDeptParentId string `json:"SourceDeptParentId"`
	SourceUserNum      int    `json:"sourceUserNum"`
}

// WeComGroupAddReq 添加企业微信资源结构体
type WeComGroupAddReq struct {
	GroupType string `json:"groupType" validate:"required,min=1,max=20"`
	GroupName string `json:"groupName" validate:"required,min=1,max=128"`
	//父级Id 大于等于0 必填
	ParentId           uint   `json:"parentId" validate:"omitempty,min=0"`
	Remark             string `json:"remark" validate:"min=0,max=128"` // 分组的中文描述
	SourceDeptId       string `json:"sourceDeptId"`
	Source             string `json:"source"`
	SourceDeptParentId string `json:"SourceDeptParentId"`
	SourceUserNum      int    `json:"sourceUserNum"`
}

// GroupUpdateReq 更新资源结构体
type GroupUpdateReq struct {
	ID        uint   `json:"id" form:"id" validate:"required"`
	GroupName string `json:"groupName" validate:"required,min=1,max=128"`
	Remark    string `json:"remark" validate:"min=0,max=128"` // 分组的中文描述
}

// GroupDeleteReq 删除资源结构体
type GroupDeleteReq struct {
	GroupIds []uint `json:"groupIds" validate:"required"`
}

// GroupGetTreeReq 获取资源树结构体
type GroupGetTreeReq struct {
	GroupName string `json:"groupName" form:"groupName"`
	Remark    string `json:"remark" form:"remark"`
	PageNum   int    `json:"pageNum" form:"pageNum"`
	PageSize  int    `json:"pageSize" form:"pageSize"`
}

type GroupAddUserReq struct {
	GroupID uint   `json:"groupId" validate:"required"`
	UserIds []uint `json:"userIds" validate:"required"`
}

type GroupRemoveUserReq struct {
	GroupID uint   `json:"groupId" validate:"required"`
	UserIds []uint `json:"userIds" validate:"required"`
}

// UserInGroupReq 在分组内的用户
type UserInGroupReq struct {
	GroupID  uint   `json:"groupId" form:"groupId" validate:"required"`
	Nickname string `json:"nickname" form:"nickname"`
}

// UserNoInGroupReq 不在分组内的用户
type UserNoInGroupReq struct {
	GroupID  uint   `json:"groupId" form:"groupId" validate:"required"`
	Nickname string `json:"nickname" form:"nickname"`
}

// SyncDingTalkDeptsReq 同步钉钉部门信息
type SyncDingTalkDeptsReq struct {
}

// SyncWeComDeptsReq 同步企业微信部门信息
type SyncWeComDeptsReq struct {
}

// SyncFeiShuDeptsReq 同步飞书部门信息
type SyncFeiShuDeptsReq struct {
}

// SyncOpenLdapDeptsReq 同步原ldap部门信息
type SyncOpenLdapDeptsReq struct {
}

// SyncOpenLdapDeptsReq 同步原ldap部门信息
type SyncSqlGrooupsReq struct {
	GroupIds []uint `json:"groupIds" validate:"required"`
}

type GroupListRsp struct {
	Total  int64   `json:"total"`
	Groups []Group `json:"groups"`
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
