package model

import (
	"encoding/base64"
	"errors"
	"fmt"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/tools"
	"strings"
	"time"

	totpModel "micro-net-hub/internal/module/totp/model"

	"github.com/patrickmn/go-cache"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/thoas/go-funk"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 当前用户信息缓存，避免频繁获取数据库
var UserInfoCache = cache.New(24*time.Hour, 48*time.Hour)

func cacheUserInfoKeygen(username string) string {
	return fmt.Sprintf("u_%s", username)
}
func CacheUserInfoGet(username string) *User {
	key := cacheUserInfoKeygen(username)
	if cache, found := UserInfoCache.Get(key); found {
		return cache.(*User)
	}
	return nil
}

func CacheUserInfoSet(username string, userInfo *User) {
	key := cacheUserInfoKeygen(username)
	UserInfoCache.Set(key, userInfo, cache.DefaultExpiration)
}

func CacheUserInfoDel(username string) {
	key := cacheUserInfoKeygen(username)
	if _, found := UserInfoCache.Get(key); found {
		UserInfoCache.Delete(key)
	}
}

func CacheUserInfoClear() {
	UserInfoCache.Flush()
}

type UserStatus uint

const (
	UserNormal   UserStatus = iota + 1 // 1 正常
	UserDisabled                       // 2 禁用
)

type UserSyncStatus uint

const (
	UserSyncNormal   UserSyncStatus = iota + 1 // 1 已同步
	UserSyncUnNormal                           // 2 未同步
)

type User struct {
	gorm.Model
	Username      string         `gorm:"type:varchar(50);not null;unique;comment:'用户名'" json:"username"` // 用户名
	Password      string         `gorm:"size:255;not null;comment:'用户密码'" json:"-"`                      // 用户密码
	Nickname      string         `gorm:"type:varchar(50);comment:'中文名'" json:"nickname"`                 // 昵称
	GivenName     string         `gorm:"type:varchar(50);comment:'花名'" json:"givenName"`                 // 花名，如果有的话，没有的话用昵称占位
	Mail          string         `gorm:"type:varchar(100);not null;unique;comment:'邮箱'" json:"mail"`     // 邮箱
	JobNumber     string         `gorm:"type:varchar(20);comment:'工号'" json:"jobNumber"`                 // 工号
	Mobile        string         `gorm:"type:varchar(15);comment:'手机号'" json:"mobile"`                   // 手机号
	Avatar        string         `gorm:"type:varchar(255);comment:'头像'" json:"avatar"`                   // 头像
	PostalAddress string         `gorm:"type:varchar(255);comment:'地址'" json:"postalAddress"`            // 地址
	Position      string         `gorm:"type:varchar(128);comment:'职位'" json:"position"`                 //  职位
	Introduction  string         `gorm:"type:varchar(255);comment:'个人简介'" json:"introduction"`           // 个人简介
	Status        UserStatus     `gorm:"type:tinyint(1);default:1;comment:'状态:1在职, 2离职'" json:"status"`  // 状态
	Creator       string         `gorm:"type:varchar(20);comment:'创建者'" json:"creator"`                  // 创建者
	Groups        []*Group       `gorm:"many2many:group_users;comment:'用户组/部门'" json:"groups"`
	Roles         []*Role        `gorm:"many2many:user_roles" json:"roles"` // 角色
	Totp          totpModel.Totp `json:"-"`
	Source        string         `gorm:"type:varchar(50);comment:'用户来源：dingTalk、wecom、feishu、ldap、platform'" json:"source"` // 来源
	SourceUserId  string         `gorm:"type:varchar(100);not null;comment:'第三方用户id'" json:"sourceUserId"`                  // 第三方用户id
	SourceUnionId string         `gorm:"type:varchar(100);not null;comment:'第三方唯一unionId'" json:"sourceUnionId"`            // 第三方唯一unionId
	UserDN        string         `gorm:"type:varchar(255);not null;comment:'用户dn'" json:"userDn"`                           // 用户在ldap的dn
	SyncState     UserSyncStatus `gorm:"type:tinyint(1);default:1;comment:'同步状态:1已同步, 2未同步'" json:"syncState"`              // 数据到ldap的同步状态
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

// 用户信息的预置处理
func (u *User) CheckAttrVacancies() {
	if u.Nickname == "" {
		u.Nickname = u.Username
	}
	if u.GivenName == "" {
		u.GivenName = u.Username
	}
	if u.Introduction == "" {
		u.Introduction = tools.ConvertBaseDNToDomain(config.Conf.Ldap.BaseDN)
	}
	// 兼容
	if u.Mail == "" || !tools.CheckEmail(u.Mail) {
		if len(config.Conf.Ldap.DefaultEmailSuffix) > 0 {
			u.Mail = u.Username + "@" + config.Conf.Ldap.DefaultEmailSuffix
		} else {
			u.Mail = u.Username + "@example.com"
		}
	}
	if u.JobNumber == "" {
		u.JobNumber = "0000"
	}
	if u.Position == "" {
		u.Position = "Default Position"
	}
	if u.PostalAddress == "" {
		u.PostalAddress = "Default PostalAddr"
	}
	if u.Mobile == "" {
		// user.Mobile = generateMobile()
		u.Mobile = "10000000000"
	}
	if tools.CheckQQNo(u.Avatar) {
		u.Avatar = fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%s&s=100", u.Avatar)
	}
}

// Update 更新资源
func (u *User) Update() error {
	if u.Password != "" {
		u.Password = tools.NewGenPasswd(u.Password)
	}

	tx := global.DB.Begin()
	err := tx.Model(u).Updates(u).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	// FIX: 临时修改 DepartmentIds 空值不更新的问题
	err = tx.Model(u).Select("DepartmentIds").Updates(u).Error
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Model(u).Association("Roles").Replace(u.Roles)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	userDb := &User{}
	global.DB.Where("username = ?", u.Username).Preload("Roles").First(&userDb)
	UserInfoCache.Set(u.Username, *userDb, cache.DefaultExpiration)

	return nil
}

// ChangePwd 更新密码
func (u *User) ChangePwd(hashNewPasswd string) error {
	err := global.DB.Model(&User{}).Where("username = ?", u.Username).Update("password", hashNewPasswd).Error
	// 如果更新密码成功，则更新当前用户信息缓存
	// 先获取缓存
	cacheUser, found := UserInfoCache.Get(u.Username)
	if err == nil {
		if found {
			user := cacheUser.(User)
			user.Password = hashNewPasswd
			UserInfoCache.Set(u.Username, user, cache.DefaultExpiration)
		} else {
			// 没有缓存就获取用户信息缓存
			var user User
			global.DB.Where("username = ?", u.Username).Preload("Roles").First(&user)
			UserInfoCache.Set(u.Username, user, cache.DefaultExpiration)
		}
	}

	return err
}

// Exist 判断资源是否存在
func (u *User) Exist(filter map[string]interface{}) bool {
	err := global.DB.Where(filter).First(&u).Error
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// Find 获取单个资源
func (u *User) Find(filter map[string]interface{}) error {
	return global.DB.Where(filter).Preload("Groups").Preload("Roles").Preload("Totp").First(&u).Error
}

// UserExistsInGroup checks if a user exists in a specific group.
func UserExistsInGroup(username, groupDN string) (bool, error) {
	var count int64
	err := global.DB.Model(&User{}). // Specify the model to avoid loading unnecessary data
						Joins("INNER JOIN group_users ON users.id = group_users.user_id").
						Joins("INNER JOIN groups ON group_users.group_id = groups.id").
						Where("users.username = ? AND groups.group_dn = ?", username, groupDN).
						Count(&count).Error // Use Count to get the number of records matching the criteria

	if err != nil {
		// Check for specific errors like connection issues or syntax errors
		if err == gorm.ErrRecordNotFound {
			return false, nil // No records found is not an error but a valid result
		}
		return false, err // Return the error to the caller for further handling
	}

	return count > 0, nil // If count is greater than 0, the user exists in the group
}

// Find 获取同名用户已入库的序号最大的用户信息
func (u *User) FindTheSameUserName(username string) error {
	return global.DB.Where("username REGEXP ? ", fmt.Sprintf("^%s[0-9]{0,3}$", username)).Order("username desc").First(&u).Error
}

func (u *User) GetQrcodestr() (qrcodestr string) {
	return fmt.Sprintf(`otpauth://totp/%s_%s?secret=%s`, strings.ReplaceAll(config.Conf.Notice.ProjectName, " ", "-"), u.Username, u.Totp.Secret)
}

func (u *User) GetRawPngBase64() (qrRawPngBase64 string, err error) {
	qrcodestr := u.GetQrcodestr()
	qrRawPng, err := qrcode.Encode(qrcodestr, qrcode.Medium, 224)
	if err != nil {
		global.Log.Errorf("%s generate qrcode error : %s", u.Username, err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(qrRawPng), nil
}

// Add 添加资源
func (u *User) Add() error {
	u.Password = tools.NewGenPasswd(u.Password)
	//result := global.DB.Create(user)
	//return user.ID, result.Error
	u.Totp.SetTotpSecret()
	return global.DB.Create(u).Error
}

// GetUserById 获取单个用户
func (u *User) GetUserById(id uint) error {
	err := global.DB.Where("id = ?", id).Preload("Roles").Preload("Totp").First(&u).Error
	return err
}

// ChangeStatus 更新用户的启用状态 和 ldap 同步状态
func (u *User) ChangeStatus(status UserStatus, syncStatus UserSyncStatus) error {
	return global.DB.Model(&User{}).Where("id = ?", u.ID).UpdateColumns(User{Status: status, SyncState: syncStatus}).Error
}

// Login 登录
func (u *User) Login() (*User, error) {
	// 根据用户名获取用户(正常状态:用户状态正常)
	var userRight User
	err := userRight.Find(map[string]interface{}{"username": u.Username})
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	// 判断用户的状态
	userStatus := userRight.Status
	if userStatus != 1 {
		return nil, errors.New("用户名或密码错误")
	}

	// 判断用户拥有的所有角色的状态,全部角色都被禁用则不能登录
	roles := userRight.Roles
	roleValid := false
	for _, role := range roles {
		// 有一个正常状态的角色就可以登录
		if role.Status == 1 {
			roleValid = true
			break
		}
	}

	if !roleValid {
		return nil, errors.New("用户角色被禁用")
	}

	if tools.NewParsePasswd(userRight.Password) != u.Password {
		return nil, errors.New("用户名或密码错误")
	}
	return &userRight, nil
}

// 当前用户信息缓存，避免频繁获取数据库
var UsersInfoCache = cache.New(24*time.Hour, 48*time.Hour)

func cacheUsersInfoKeygen(groupdn string) string {
	if groupdn == "" {
		return "us_all"
	}
	gdn := strings.ReplaceAll(groupdn, ",", "_")
	gdn = strings.ReplaceAll(gdn, "=", "-")
	return fmt.Sprintf("us_%s", gdn)
}

func CacheUsersInfoGet(groupdn string) *Users {
	key := cacheUsersInfoKeygen(groupdn)
	if cache, found := UsersInfoCache.Get(key); found {
		return cache.(*Users)
	}
	return nil
}

func CacheUsersInfoSet(groupdn string, usersInfo *Users) {
	key := cacheUsersInfoKeygen(groupdn)
	UsersInfoCache.Set(key, usersInfo, cache.DefaultExpiration)
}

func CacheUsersInfoDel(groupdn string) {
	key := cacheUsersInfoKeygen(groupdn)
	if _, found := UsersInfoCache.Get(key); found {
		UsersInfoCache.Delete(key)
	}
}

func CacheUsersInfoClear() {
	UsersInfoCache.Flush()
}

type Users []*User

func NewUsers() Users {
	return make(Users, 0)
}

// List 获取数据列表
func (us *Users) List(req *User, pageNum int, pageSize int) error {
	db := global.DB.Model(&User{}).Order("id DESC")

	username := strings.TrimSpace(req.Username)
	if username != "" {
		db = db.Where("username LIKE ?", fmt.Sprintf("%%%s%%", username))
	}
	nickname := strings.TrimSpace(req.Nickname)
	if nickname != "" {
		db = db.Where("nickname LIKE ?", fmt.Sprintf("%%%s%%", nickname))
	}
	mobile := strings.TrimSpace(req.Mobile)
	if mobile != "" {
		db = db.Where("mobile LIKE ?", fmt.Sprintf("%%%s%%", mobile))
	}
	givenName := strings.TrimSpace(req.GivenName)
	if givenName != "" {
		db = db.Where("given_name LIKE ?", fmt.Sprintf("%%%s%%", givenName))
	}
	status := req.Status
	if status != 0 {
		db = db.Where("status = ?", status)
	}
	syncState := req.SyncState
	if syncState != 0 {
		db = db.Where("sync_state = ?", syncState)
	}

	pageReq := tools.NewPageOption(pageNum, pageSize)
	err := db.Offset(pageReq.PageNum).Limit(pageReq.PageSize).Preload("Roles").Preload("Groups").Find(&us).Error
	return err
}

// List 获取数据列表
func (us *Users) ListAll() (err error) {
	err = global.DB.Model(&User{}).Order("created_at DESC").Find(&us).Error
	return err
}

// Find 获取多个 User
func (us *Users) GetUsersByGroupDN(groupDN string) error {
	return global.DB.Model(&User{}). // Specify the model to avoid loading unnecessary data
						Joins("INNER JOIN group_users ON users.id = group_users.user_id").
						Joins("INNER JOIN groups ON group_users.group_id = groups.id").
						Where("groups.group_dn = ?", groupDN).
						Find(&us).Error
}

// GetUserByIds 根据用户ID获取用户角色排序最小值
func (us *Users) GetUsersByIds(ids []uint) error {
	// 根据用户ID获取用户信息
	err := global.DB.Where("id IN (?)", ids).Preload("Groups").Preload("Roles").Find(&us).Error
	return err
}

// Delete 批量删除
func DeleteUsersById(ids []uint) error {
	// 用户和角色存在多对多关联关系
	var us = NewUsers()
	for _, id := range ids {
		// 根据ID获取用户
		filter := map[string]interface{}{"id": id}

		user := new(User)
		err := user.Find(filter)
		if err != nil {
			return fmt.Errorf("需要删除的用户获取信息失败，err: %v", err)
		}
		us = append(us, user)
	}

	err := global.DB.Select(clause.Associations).Unscoped().Delete(&us).Error
	if err != nil {
		return err
	} else {
		// 删除用户成功，则删除用户信息缓存
		for _, u := range us {
			UserInfoCache.Delete(u.Username)
		}
	}

	// 删除用户在group的关联
	err = global.DB.Exec("DELETE FROM group_users WHERE user_id IN (?)", ids).Error
	if err != nil {
		return err
	}

	return err
}

// GetUserMinRoleSortsByIds 根据用户ID获取用户角色排序最小值
func GetUserMinRoleSortsByIds(ids []uint) ([]int, error) {
	// 根据用户ID获取用户信息
	var us = NewUsers()
	err := global.DB.Where("id IN (?)", ids).Preload("Roles").Find(&us).Error
	if err != nil {
		return []int{}, err
	}
	if len(us) == 0 {
		return []int{}, errors.New("未获取到任何用户信息")
	}
	var roleMinSortList []int
	for _, user := range us {
		roles := user.Roles
		var roleSortList []int
		for _, role := range roles {
			roleSortList = append(roleSortList, int(role.Sort))
		}
		roleMinSort := funk.MinInt(roleSortList)
		roleMinSortList = append(roleMinSortList, roleMinSort)
	}
	return roleMinSortList, nil
}

// Count 获取数据总数
func UserCount() (int64, error) {
	var count int64
	err := global.DB.Model(&User{}).Count(&count).Error
	return count, err
}

// ListCout 获取符合条件的数据列表条数
func UserListCount(req *User) (int64, error) {
	var count int64
	db := global.DB.Model(&User{}).Order("id DESC")

	username := strings.TrimSpace(req.Username)
	if username != "" {
		db = db.Where("username LIKE ?", fmt.Sprintf("%%%s%%", username))
	}
	nickname := strings.TrimSpace(req.Nickname)
	if nickname != "" {
		db = db.Where("nickname LIKE ?", fmt.Sprintf("%%%s%%", nickname))
	}
	mobile := strings.TrimSpace(req.Mobile)
	if mobile != "" {
		db = db.Where("mobile LIKE ?", fmt.Sprintf("%%%s%%", mobile))
	}
	givenName := strings.TrimSpace(req.GivenName)
	if givenName != "" {
		db = db.Where("given_name LIKE ?", fmt.Sprintf("%%%s%%", givenName))
	}
	status := req.Status
	if status != 0 {
		db = db.Where("status = ?", status)
	}
	syncState := req.SyncState
	if syncState != 0 {
		db = db.Where("sync_state = ?", syncState)
	}

	err := db.Count(&count).Error
	return count, err
}
