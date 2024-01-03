package model

import (
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	Name    string  `gorm:"type:varchar(20);not null;unique" json:"name"`
	Keyword string  `gorm:"type:varchar(20);not null;unique" json:"keyword"`
	Remark  string  `gorm:"type:varchar(100);comment:'备注'" json:"remark"`
	Status  uint    `gorm:"type:tinyint(1);default:1;comment:'1正常, 2禁用'" json:"status"`
	Sort    uint    `gorm:"type:int(3);default:999;comment:'角色排序(排序越大权限越低, 不能查看比自己序号小的角色, 不能编辑同序号用户权限, 排序为1表示超级管理员)'" json:"sort"`
	Creator string  `gorm:"type:varchar(20);" json:"creator"`
	Users   []*User `gorm:"many2many:user_roles" json:"users"`
	Menus   []*Menu `gorm:"many2many:role_menus;" json:"menus"` // 角色菜单多对多关系
}

// RoleAddReq 添加资源结构体
type RoleAddReq struct {
	Name    string `json:"name" validate:"required,min=1,max=20"`
	Keyword string `json:"keyword" validate:"required,min=1,max=20"`
	Remark  string `json:"remark" validate:"min=0,max=100"`
	Status  uint   `json:"status" validate:"oneof=1 2"`
	Sort    uint   `json:"sort" validate:"gte=1,lte=999"`
}

// RoleListReq 列表结构体
type RoleListReq struct {
	Name     string `json:"name" form:"name"`
	Keyword  string `json:"keyword" form:"keyword"`
	Status   uint   `json:"status" form:"status"`
	PageNum  int    `json:"pageNum" form:"pageNum"`
	PageSize int    `json:"pageSize" form:"pageSize"`
}

// RoleUpdateReq 更新资源结构体
type RoleUpdateReq struct {
	ID      uint   `json:"id" validate:"required"`
	Name    string `json:"name" validate:"required,min=1,max=20"`
	Keyword string `json:"keyword" validate:"required,min=1,max=20"`
	Remark  string `json:"remark" validate:"min=0,max=100"`
	Status  uint   `json:"status" validate:"oneof=1 2"`
	Sort    uint   `json:"sort" validate:"gte=1,lte=999"`
}

// RoleDeleteReq 删除资源结构体
type RoleDeleteReq struct {
	RoleIds []uint `json:"roleIds" validate:"required"`
}

// RoleGetTreeReq 获取资源树结构体
type RoleGetTreeReq struct {
}

// RoleGetMenuListReq 获取角色菜单列表结构体
type RoleGetMenuListReq struct {
	RoleID uint `json:"roleId" form:"roleId" validate:"required"`
}

// RoleGetApiListReq 获取角色接口列表结构体
type RoleGetApiListReq struct {
	RoleID uint `json:"roleId" form:"roleId" validate:"required"`
}

// RoleUpdateMenusReq 更新角色菜单结构体
type RoleUpdateMenusReq struct {
	RoleID  uint   `json:"roleId" validate:"required"`
	MenuIds []uint `json:"menuIds" validate:"required"`
}

// RoleUpdateApisReq 更新角色接口结构体
type RoleUpdateApisReq struct {
	RoleID uint   `json:"roleId" validate:"required"`
	ApiIds []uint `json:"apiIds" validate:"required"`
}

type RoleListRsp struct {
	Total int64  `json:"total"`
	Roles []Role `json:"roles"`
}
