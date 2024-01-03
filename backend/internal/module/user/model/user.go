package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username      string  `gorm:"type:varchar(50);not null;unique;comment:'用户名'" json:"username"`                    // 用户名
	Password      string  `gorm:"size:255;not null;comment:'用户密码'" json:"password"`                                  // 用户密码
	Nickname      string  `gorm:"type:varchar(50);comment:'中文名'" json:"nickname"`                                    // 昵称
	GivenName     string  `gorm:"type:varchar(50);comment:'花名'" json:"givenName"`                                    // 花名，如果有的话，没有的话用昵称占位
	Mail          string  `gorm:"type:varchar(100);comment:'邮箱'" json:"mail"`                                        // 邮箱
	JobNumber     string  `gorm:"type:varchar(20);comment:'工号'" json:"jobNumber"`                                    // 工号
	Mobile        string  `gorm:"type:varchar(15);not null;unique;comment:'手机号'" json:"mobile"`                      // 手机号
	Avatar        string  `gorm:"type:varchar(255);comment:'头像'" json:"avatar"`                                      // 头像
	PostalAddress string  `gorm:"type:varchar(255);comment:'地址'" json:"postalAddress"`                               // 地址
	Departments   string  `gorm:"type:varchar(512);comment:'部门'" json:"departments"`                                 // 部门
	Position      string  `gorm:"type:varchar(128);comment:'职位'" json:"position"`                                    //  职位
	Introduction  string  `gorm:"type:varchar(255);comment:'个人简介'" json:"introduction"`                              // 个人简介
	Status        uint    `gorm:"type:tinyint(1);default:1;comment:'状态:1在职, 2离职'" json:"status"`                     // 状态
	Creator       string  `gorm:"type:varchar(20);;comment:'创建者'" json:"creator"`                                    // 创建者
	Source        string  `gorm:"type:varchar(50);comment:'用户来源：dingTalk、wecom、feishu、ldap、platform'" json:"source"` // 来源
	DepartmentId  string  `gorm:"type:varchar(100);not null;comment:'部门id'" json:"departmentId"`                     // 部门id
	Roles         []*Role `gorm:"many2many:user_roles" json:"roles"`                                                 // 角色
	SourceUserId  string  `gorm:"type:varchar(100);not null;comment:'第三方用户id'" json:"sourceUserId"`                  // 第三方用户id
	SourceUnionId string  `gorm:"type:varchar(100);not null;comment:'第三方唯一unionId'" json:"sourceUnionId"`            // 第三方唯一unionId
	UserDN        string  `gorm:"type:varchar(255);not null;comment:'用户dn'" json:"userDn"`                           // 用户在ldap的dn
	SyncState     uint    `gorm:"type:tinyint(1);default:1;comment:'同步状态:1已同步, 2未同步'" json:"syncState"`              // 数据到ldap的同步状态
}

func (u *User) SetUserName(userName string) {
	u.Username = userName
}

func (u *User) SetNickName(nickName string) {
	u.Nickname = nickName
}

func (u *User) SetGivenName(givenName string) {
	u.GivenName = givenName
}

func (u *User) SetMail(mail string) {
	u.Mail = mail
}

func (u *User) SetJobNumber(jobNum string) {
	u.JobNumber = jobNum
}

func (u *User) SetMobile(mobile string) {
	u.Mobile = mobile
}

func (u *User) SetAvatar(avatar string) {
	u.Avatar = avatar
}

func (u *User) SetPostalAddress(address string) {
	u.PostalAddress = address
}

func (u *User) SetPosition(position string) {
	u.Position = position
}

func (u *User) SetIntroduction(desc string) {
	u.Introduction = desc
}

func (u *User) SetSourceUserId(sourceUserId string) {
	u.SourceUserId = sourceUserId
}

func (u *User) SetSourceUnionId(sourceUnionId string) {
	u.SourceUnionId = sourceUnionId
}

// UserAddReq 创建资源结构体
type UserAddReq struct {
	Username      string `json:"username" validate:"required,min=2,max=50"`
	Password      string `json:"password"`
	Nickname      string `json:"nickname" validate:"required,min=0,max=50"`
	GivenName     string `json:"givenName" validate:"min=0,max=50"`
	Mail          string `json:"mail" validate:"required,min=0,max=100"`
	JobNumber     string `json:"jobNumber" validate:"required,min=0,max=20"`
	PostalAddress string `json:"postalAddress" validate:"min=0,max=255"`
	Departments   string `json:"departments" validate:"min=0,max=512"`
	Position      string `json:"position" validate:"min=0,max=128"`
	Mobile        string `json:"mobile" validate:"required,checkMobile"`
	Avatar        string `json:"avatar"`
	Introduction  string `json:"introduction" validate:"min=0,max=255"`
	Status        uint   `json:"status" validate:"oneof=1 2"`
	DepartmentId  []uint `json:"departmentId" validate:"required"`
	Source        string `json:"source" validate:"min=0,max=50"`
	RoleIds       []uint `json:"roleIds" validate:"required"`
}

// DingUserAddReq 钉钉用户创建资源结构体
type DingUserAddReq struct {
	Username      string `json:"username" validate:"required,min=2,max=50"`
	Password      string `json:"password"`
	Nickname      string `json:"nickname" validate:"required,min=0,max=50"`
	GivenName     string `json:"givenName" validate:"min=0,max=50"`
	Mail          string `json:"mail" validate:"required,min=0,max=100"`
	JobNumber     string `json:"jobNumber" validate:"required,min=0,max=20"`
	PostalAddress string `json:"postalAddress" validate:"min=0,max=255"`
	Departments   string `json:"departments" validate:"min=0,max=512"`
	Position      string `json:"position" validate:"min=0,max=128"`
	Mobile        string `json:"mobile" validate:"required,checkMobile"`
	Avatar        string `json:"avatar"`
	Introduction  string `json:"introduction" validate:"min=0,max=255"`
	Status        uint   `json:"status" validate:"oneof=1 2"`
	DepartmentId  []uint `json:"departmentId" validate:"required"`
	Source        string `json:"source" validate:"min=0,max=50"`
	RoleIds       []uint `json:"roleIds" validate:"required"`
	SourceUserId  string `json:"sourceUserId"`  // 第三方用户id
	SourceUnionId string `json:"sourceUnionId"` // 第三方唯一unionId
}

// WeComUserAddReq 企业微信用户创建资源结构体
type WeComUserAddReq struct {
	Username      string `json:"username" validate:"required,min=2,max=50"`
	Password      string `json:"password"`
	Nickname      string `json:"nickname" validate:"required,min=0,max=50"`
	GivenName     string `json:"givenName" validate:"min=0,max=50"`
	Mail          string `json:"mail" validate:"required,min=0,max=100"`
	JobNumber     string `json:"jobNumber" validate:"required,min=0,max=20"`
	PostalAddress string `json:"postalAddress" validate:"min=0,max=255"`
	Departments   string `json:"departments" validate:"min=0,max=512"`
	Position      string `json:"position" validate:"min=0,max=128"`
	Mobile        string `json:"mobile" validate:"required,checkMobile"`
	Avatar        string `json:"avatar"`
	Introduction  string `json:"introduction" validate:"min=0,max=255"`
	Status        uint   `json:"status" validate:"oneof=1 2"`
	DepartmentId  []uint `json:"departmentId" validate:"required"`
	Source        string `json:"source" validate:"min=0,max=50"`
	RoleIds       []uint `json:"roleIds" validate:"required"`
	SourceUserId  string `json:"sourceUserId"`  // 第三方用户id
	SourceUnionId string `json:"sourceUnionId"` // 第三方唯一unionId
}

// UserUpdateReq 更新资源结构体
type UserUpdateReq struct {
	ID            uint   `json:"id" validate:"required"`
	Username      string `json:"username" validate:"required,min=2,max=50"`
	Nickname      string `json:"nickname" validate:"min=0,max=20"`
	GivenName     string `json:"givenName" validate:"min=0,max=50"`
	Mail          string `json:"mail" validate:"min=0,max=100"`
	JobNumber     string `json:"jobNumber" validate:"min=0,max=20"`
	PostalAddress string `json:"postalAddress" validate:"min=0,max=255"`
	Departments   string `json:"departments" validate:"min=0,max=512"`
	Position      string `json:"position" validate:"min=0,max=128"`
	Mobile        string `json:"mobile" validate:"checkMobile"`
	Avatar        string `json:"avatar"`
	Introduction  string `json:"introduction" validate:"min=0,max=255"`
	DepartmentId  []uint `json:"departmentId" validate:"required"`
	Source        string `json:"source" validate:"min=0,max=50"`
	RoleIds       []uint `json:"roleIds" validate:"required"`
}

// UserDeleteReq 批量删除资源结构体
type UserDeleteReq struct {
	UserIds []uint `json:"userIds" validate:"required"`
}

// UserChangePwdReq 修改密码结构体
type UserChangePwdReq struct {
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
}

// UserChangeUserStatusReq 修改用户状态结构体
type UserChangeUserStatusReq struct {
	ID     uint `json:"id" validate:"required"`
	Status uint `json:"status" validate:"oneof=1 2"`
}

// UserGetUserInfoReq 获取用户信息结构体
type UserGetUserInfoReq struct {
}

// SyncDingUserReq 同步钉钉用户信息
type SyncDingUserReq struct {
}

// SyncWeComUserReq 同步企业微信用户信息
type SyncWeComUserReq struct {
}

// SyncFeiShuUserReq 同步飞书用户信息
type SyncFeiShuUserReq struct {
}

// SyncOpenLdapUserReq 同步ldap用户信息
type SyncOpenLdapUserReq struct {
}
type SyncSqlUserReq struct {
	UserIds []uint `json:"userIds" validate:"required"`
}

// UserListReq 获取用户列表结构体
type UserListReq struct {
	Username     string `json:"username" form:"username"`
	Mobile       string `json:"mobile" form:"mobile" `
	Nickname     string `json:"nickname" form:"nickname"`
	GivenName    string `json:"givenName" form:"givenName"`
	DepartmentId []uint `json:"departmentId" form:"departmentId"`
	Status       uint   `json:"status" form:"status" `
	SyncState    uint   `json:"syncState" form:"syncState" `
	PageNum      int    `json:"pageNum" form:"pageNum"`
	PageSize     int    `json:"pageSize" form:"pageSize"`
}

// RegisterAndLoginReq 用户登录结构体
type RegisterAndLoginReq struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type UserListRsp struct {
	Total int    `json:"total"`
	Users []User `json:"users"`
}
