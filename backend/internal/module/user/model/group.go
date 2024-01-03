package model

import (
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

// GroupListReq 获取资源列表结构体
type GroupListReq struct {
	GroupName string `json:"groupName" form:"groupName"`
	Remark    string `json:"remark" form:"remark"`
	PageNum   int    `json:"pageNum" form:"pageNum"`
	PageSize  int    `json:"pageSize" form:"pageSize"`
	SyncState uint   `json:"syncState" form:"syncState"`
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
