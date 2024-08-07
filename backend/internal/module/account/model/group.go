package model

import (
	"errors"
	"fmt"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/tools"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

type GroupSyncStatus uint

const (
	GroupSyncNormal   GroupSyncStatus = iota + 1 // 1 已同步
	GroupSyncUnNormal                            // 2 未同步
)

// 当前用户信息缓存，避免频繁获取数据库
var GroupInfoCache = cache.New(24*time.Hour, 48*time.Hour)

func cacheGroupInfoKeygen(groupname string) string {
	return fmt.Sprintf("g_%s", groupname)
}

func CacheGroupInfoGet(groupname string) *Group {
	key := cacheGroupInfoKeygen(groupname)
	if cache, found := GroupInfoCache.Get(key); found {
		return cache.(*Group)
	}
	return nil
}

func CacheGroupInfoSet(groupname string, groupInfo *Group) {
	key := cacheGroupInfoKeygen(groupname)
	GroupInfoCache.Set(key, groupInfo, cache.DefaultExpiration)
}

func CacheGroupInfoDel(groupname string) {
	key := cacheGroupInfoKeygen(groupname)
	if _, found := GroupInfoCache.Get(key); found {
		GroupInfoCache.Delete(key)
	}
}

func CacheGroupInfoClear() {
	GroupInfoCache.Flush()
}

type Group struct {
	gorm.Model
	GroupName          string          `gorm:"type:varchar(128);not null;unique;comment:'分组名称'" json:"groupName"`
	Remark             string          `gorm:"type:varchar(128);comment:'分组中文说明'" json:"remark"`
	Creator            string          `gorm:"type:varchar(20);comment:'创建人'" json:"creator"`
	ParentId           uint            `gorm:"default:0;comment:'父组编号(编号为0时表示根组)'" json:"parentId"`
	Children           []*Group        `gorm:"-" json:"children"`
	Users              []*User         `gorm:"many2many:group_users" json:"users"`
	GroupType          string          `gorm:"type:varchar(20);comment:'分组类型：cn、ou'" json:"groupType"`
	Source             string          `gorm:"type:varchar(20);comment:'来源：dingTalk、weCom、ldap、platform'" json:"source"`
	SourceDeptId       string          `gorm:"type:varchar(100);comment:'部门编号'" json:"sourceDeptId"`
	SourceDeptParentId string          `gorm:"type:varchar(100);comment:'父部门编号'" json:"sourceDeptParentId"`
	SourceUserNum      int             `gorm:"default:0;comment:'部门下的用户数量，从第三方获取的数据'" json:"source_user_num"`
	GroupDN            string          `gorm:"type:varchar(255);not null;comment:'分组dn'" json:"groupDn"`             // 分组在ldap的dn
	SyncState          GroupSyncStatus `gorm:"type:tinyint(1);default:1;comment:'同步状态:1已同步, 2未同步'" json:"syncState"` // 数据到ldap的同步状态
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
	err := global.DB.Order("created_at DESC").Where(filter).First(&g).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// AddUserToGroup 添加用户到分组
func (g *Group) AddUserToGroup(user *User) error {
	return global.DB.Model(&g).Association("Users").Append(user)
}

// RemoveUserFromGroup 将用户从分组移除
func (g *Group) RemoveUserFromGroup(user *User) error {
	return global.DB.Model(&g).Association("Users").Delete(user)
}

// AddUserToGroup 添加用户到分组
func (g *Group) AddUsersToGroup(users *Users) error {
	return global.DB.Model(&g).Association("Users").Append(users)
}

// RemoveUserFromGroup 将用户从分组移除
func (g *Group) RemoveUsersFromGroup(users *Users) error {
	return global.DB.Model(&g).Association("Users").Delete(users)
}

// GroupCount 获取数据总数
func GroupCount() (int64, error) {
	var count int64
	err := global.DB.Model(&Group{}).Count(&count).Error
	return count, err
}

// LdapDeptIdsToGroups 将 三方账号系统 部门id 转换为MySQL分组id
func ThirdPartDeptIdsToGroups(ids []string) (groups []*Group, err error) {
	err = global.DB.Model(&Group{}).Where("source_dept_id IN (?)", ids).Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return
}

// LdapDeptDnsToGroups 将 Ldap DN 转换为MySQL分组id
func LdapDeptDnsToGroups(dns []string) (groups []*Group, err error) {
	err = global.DB.Model(&Group{}).Where("group_dn IN (?)", dns).Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return
}

// 当前用户信息缓存，避免频繁获取数据库
var GroupsInfoCache = cache.New(24*time.Hour, 48*time.Hour)

func cacheGroupsInfoKeygen() string {
	return "gs_all"
}

func CacheGroupsInfoGet() Groups {
	key := cacheGroupsInfoKeygen()
	if cache, found := GroupsInfoCache.Get(key); found {
		return cache.(Groups)
	}
	return nil
}

func CacheGroupsInfoSet(groupsInfo Groups) {
	key := cacheGroupsInfoKeygen()
	GroupsInfoCache.Set(key, groupsInfo, cache.DefaultExpiration)
}

func CacheGroupsInfoDel() {
	key := cacheGroupsInfoKeygen()
	if _, found := GroupInfoCache.Get(key); found {
		GroupsInfoCache.Delete(key)
	}
}

func CacheGroupsInfoClear() {
	GroupsInfoCache.Flush()
}

type Groups []*Group

func NewGroups() Groups {
	return make([]*Group, 0)
}

func (gs *Groups) Find(filter map[string]interface{}, args ...interface{}) error {
	return global.DB.Where(filter, args).Find(&gs).Error
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
		if err := global.DB.Unscoped().Delete(&g).Error; err != nil {
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

// List 获取数据列表
func (gs *Groups) List(req *Group, pageNum int, pageSize int) error {
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

	pageReq := tools.NewPageOption(pageNum, pageSize)
	err := db.Offset(pageReq.PageNum).Limit(pageReq.PageSize).Preload("Users").Find(&list).Error
	return err
}

// List 获取数据列表
func (gs *Groups) ListTree(req *Group, pageNum int, pageSize int) error {
	db := global.DB.Model(&Group{}).Order("created_at DESC")

	groupName := strings.TrimSpace(req.GroupName)
	if groupName != "" {
		db = db.Where("group_name LIKE ?", fmt.Sprintf("%%%s%%", groupName))
	}
	groupRemark := strings.TrimSpace(req.Remark)
	if groupRemark != "" {
		db = db.Where("remark LIKE ?", fmt.Sprintf("%%%s%%", groupRemark))
	}

	pageReq := tools.NewPageOption(pageNum, pageSize)
	err := db.Offset(pageReq.PageNum).Limit(pageReq.PageSize).Find(&gs).Error
	return err
}
