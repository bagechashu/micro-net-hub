package user

import (
	"fmt"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/module/account/auth"
	"micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/module/goldap/ldapmgr"
	totpModel "micro-net-hub/internal/module/totp/model"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"
)

// UserAddReq 创建资源结构体
type UserAddReq struct {
	Username      string           `json:"username" validate:"required,min=2,max=50"`
	Password      string           `json:"password"`
	Nickname      string           `json:"nickname" validate:"required,min=0,max=50"`
	GivenName     string           `json:"givenName" validate:"min=0,max=50"`
	Mail          string           `json:"mail" validate:"required,min=0,max=100"`
	JobNumber     string           `json:"jobNumber" validate:"min=0,max=20"`
	PostalAddress string           `json:"postalAddress" validate:"min=0,max=255"`
	Position      string           `json:"position" validate:"min=0,max=128"`
	Mobile        string           `json:"mobile" validate:"checkMobile"`
	Avatar        string           `json:"avatar"`
	Introduction  string           `json:"introduction" validate:"min=0,max=255"`
	Status        model.UserStatus `json:"status" validate:"oneof=1 2"`
	GroupIds      []uint           `json:"groupIds" validate:"required"`
	Source        string           `json:"source" validate:"min=0,max=50"`
	RoleIds       []uint           `json:"roleIds" validate:"required"`
	Notice        bool             `json:"notice" validate:"omitempty"`
}

// Add 添加记录
func Add(c *gin.Context) {
	var req UserAddReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	var u model.User
	if u.Exist(map[string]interface{}{"username": req.Username}) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("用户名已存在,请勿重复添加")))
		return
	}
	// if u.Exist(map[string]interface{}{"mobile": r.Mobile}) {
	// 	helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("手机号已存在,请勿重复添加")))
	// 	return
	// }
	// if u.Exist(map[string]interface{}{"job_number": r.JobNumber}) {
	// 	helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("工号已存在,请勿重复添加")))
	// 	return
	// }
	if u.Exist(map[string]interface{}{"mail": req.Mail}) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("邮箱已存在,请勿重复添加")))
		return
	}

	// 密码通过RSA解密
	// 密码不为空就解密
	if req.Password != "" {
		decodeData, err := tools.RSADecrypt([]byte(req.Password), config.Conf.System.RSAPrivateBytes)
		if err != nil {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("密码解密失败")))
			return
		}
		req.Password = string(decodeData)
	} else {
		req.Password = tools.GeneratePassword(8)
	}

	// 当前登陆用户角色排序最小值（最高等级角色）以及当前登陆的用户
	currentRoleSortMin, ctxUser, err := auth.GetCtxLoginUserMinRole(c)
	if err != nil {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("获取当前登陆用户角色排序最小值失败")))
		return
	}

	// 根据角色id获取角色
	if req.RoleIds == nil || len(req.RoleIds) == 0 {
		req.RoleIds = []uint{2} // 默认添加为普通用户角色
	}

	roles := model.NewRoles()
	err = roles.GetRolesByIds(req.RoleIds)
	if err != nil {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败")))
		return
	}

	var reqRoleSorts []int
	for _, role := range roles {
		reqRoleSorts = append(reqRoleSorts, int(role.Sort))
	}
	// 前端传来用户角色排序最小值（最高等级角色）
	reqRoleSortMin := uint(funk.MinInt(reqRoleSorts))

	// 如果登录用户的角色ID为1，亦即为管理员，则直接放行，保障管理员拥有最大权限
	if currentRoleSortMin != 1 {
		// 当前用户的角色排序最小值 需要小于 前端传来的角色排序最小值（用户不能创建比自己等级高的或者相同等级的用户）
		if currentRoleSortMin >= reqRoleSortMin {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("用户不能创建比自己等级高的或者相同等级的用户")))
			return
		}
	}
	user := model.User{
		Username:      req.Username,
		Password:      req.Password,
		Nickname:      req.Nickname,
		GivenName:     req.GivenName,
		Mail:          req.Mail,
		JobNumber:     req.JobNumber,
		Mobile:        req.Mobile,
		Avatar:        req.Avatar,
		PostalAddress: req.PostalAddress,
		Position:      req.Position,
		Introduction:  req.Introduction,
		Status:        req.Status,
		Creator:       ctxUser.Username,
		Source:        req.Source,
		Roles:         roles,
		UserDN:        fmt.Sprintf("uid=%s,%s", req.Username, config.Conf.Ldap.UserDN),
	}

	if user.Source == "" {
		user.Source = "platform"
	}

	// 获取用户将要添加的分组

	var gs = model.NewGroups()
	err = gs.GetGroupsByIds(req.GroupIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败: "+err.Error())))
		return
	}
	user.Groups = gs
	err = CommonAddUser(&user)
	if err != nil {
		helper.ErrV2(c, helper.NewOperationError(fmt.Errorf("添加用户失败: "+err.Error())))
		return
	}

	if req.Notice {
		var nu model.User
		if nu.Find(map[string]interface{}{"username": req.Username}) != nil {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("系统通知用户账号信息失败, 请手工通知")))
			return
		}
		qrRawPngBase64, err := nu.GetRawPngBase64()
		if err != nil {
			helper.ErrV2(c, helper.NewOperationError(fmt.Errorf("系统通知用户账号信息失败, 请手工通知")))
			return
		}
		if err := tools.SendUserInfo([]string{nu.Mail}, nu.Username, req.Password, qrRawPngBase64); err != nil {
			helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("邮件发送新用户账户信息失败, 请手工通知"+err.Error())))
			return
		}
	}

	helper.Success(c, nil)
}

// UserUpdateReq 更新资源结构体
type UserUpdateReq struct {
	ID            uint   `json:"id" validate:"required"`
	Username      string `json:"username" validate:"required,min=2,max=50"`
	Password      string `json:"password"`
	Nickname      string `json:"nickname" validate:"min=0,max=20"`
	GivenName     string `json:"givenName" validate:"min=0,max=50"`
	Mail          string `json:"mail" validate:"min=0,max=100"`
	JobNumber     string `json:"jobNumber" validate:"min=0,max=20"`
	PostalAddress string `json:"postalAddress" validate:"min=0,max=255"`
	Position      string `json:"position" validate:"min=0,max=128"`
	Mobile        string `json:"mobile" validate:"checkMobile"`
	Avatar        string `json:"avatar"`
	Introduction  string `json:"introduction" validate:"min=0,max=255"`
	GroupIds      []uint `json:"groupIds" validate:"required"`
	Source        string `json:"source" validate:"min=0,max=50"`
	RoleIds       []uint `json:"roleIds" validate:"required"`
	Notice        bool   `json:"notice" validate:"omitempty"`
}

// Update 更新记录
func Update(c *gin.Context) {
	var req UserUpdateReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	var u model.User
	if !u.Exist(map[string]interface{}{"id": req.ID}) {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("该记录不存在")))
		return
	}

	// 获取当前登陆用户
	ctxUser, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败")))
		return
	}

	// 获取当前登陆用户角色ID集合
	var currentRoleSorts []int
	for _, role := range ctxUser.Roles {
		currentRoleSorts = append(currentRoleSorts, int(role.Sort))
	}

	// 获取将要操作的用户角色ID集合
	var reqRoleSorts []int

	roles := model.NewRoles()
	err = roles.GetRolesByIds(req.RoleIds)
	if err != nil {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败")))
		return
	}
	if len(roles) == 0 {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败")))
		return
	}
	for _, role := range roles {
		reqRoleSorts = append(reqRoleSorts, int(role.Sort))
	}

	// 当前登陆用户角色排序最小值（最高等级角色）
	currentRoleSortMin := funk.MinInt(currentRoleSorts)
	// 前端传来用户角色排序最小值（最高等级角色）
	reqRoleSortMin := funk.MinInt(reqRoleSorts)

	// 如果登录用户的角色ID为1，亦即为管理员，则直接放行，保障管理员拥有最大权限
	if currentRoleSortMin != model.SuperAdminRoleID {
		// 判断是更新自己还是更新别人,如果操作的ID与登陆用户的ID一致，则说明操作的是自己
		if int(req.ID) == int(ctxUser.ID) {
			// 不能更改自己的角色
			reqDiff, currentDiff := funk.Difference(reqRoleSorts, currentRoleSorts)
			if len(reqDiff.([]int)) > 0 || len(currentDiff.([]int)) > 0 {
				helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("不能更改自己的角色")))
				return
			}
		}

		// 如果是更新别人，操作者不能更新比自己角色等级高的或者相同等级的用户
		minRoleSorts, err := model.GetUserMinRoleSortsByIds([]uint{uint(req.ID)}) // 根据userIdID获取用户角色排序最小值
		if err != nil || len(minRoleSorts) == 0 {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("根据用户ID获取用户角色排序最小值失败")))
			return
		}
		if currentRoleSortMin >= minRoleSorts[0] || currentRoleSortMin >= reqRoleSortMin {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("用户不能更新比自己角色等级高的或者相同等级的用户")))
			return
		}
	}

	// 先获取用户信息
	oldData := new(model.User)
	err = oldData.Find(map[string]interface{}{"id": req.ID})
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(err))
		return
	}

	// 拼装新的用户信息
	user := model.User{
		Model:         oldData.Model,
		Username:      req.Username,
		Nickname:      req.Nickname,
		GivenName:     req.GivenName,
		Mail:          req.Mail,
		JobNumber:     req.JobNumber,
		Mobile:        req.Mobile,
		Avatar:        req.Avatar,
		PostalAddress: req.PostalAddress,
		Position:      req.Position,
		Introduction:  req.Introduction,
		Creator:       ctxUser.Username,
		Source:        oldData.Source,
		Roles:         roles,
		UserDN:        oldData.UserDN,
	}

	// 密码不为空就解密并更新, 为空则不更新
	if req.Password != "" {
		decodeData, err := tools.RSADecrypt([]byte(req.Password), config.Conf.System.RSAPrivateBytes)
		if err != nil {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("密码解密失败")))
			return
		}
		user.Password = string(decodeData)
	}

	if err = CommonUpdateUser(oldData, &user, req.GroupIds); err != nil {
		helper.ErrV2(c, helper.NewOperationError(fmt.Errorf("更新用户失败: "+err.Error())))
		return
	}

	// 删除缓存
	model.CacheUserInfoDel(oldData.Username)

	if req.Notice {
		var nu model.User
		if nu.Find(map[string]interface{}{"username": req.Username}) != nil {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("系统通知用户账号信息失败, 请手工通知")))
			return
		}
		qrRawPngBase64, err := nu.GetRawPngBase64()
		if err != nil {
			helper.ErrV2(c, helper.NewOperationError(fmt.Errorf("系统通知用户账号信息失败, 请手工通知")))
			return
		}
		if user.Password == "" {
			user.Password = "Use the original password"
		}
		if err := tools.SendUserInfo([]string{nu.Mail}, nu.Username, user.Password, qrRawPngBase64); err != nil {
			helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("邮件发送用户账号更新信息失败, 请手工通知"+err.Error())))
			return
		}
	}

	helper.Success(c, nil)
}

// UserListReq 获取用户列表结构体
type UserListReq struct {
	Username  string               `json:"username" form:"username"`
	Nickname  string               `json:"nickname" form:"nickname"`
	Status    model.UserStatus     `json:"status" form:"status" `
	SyncState model.UserSyncStatus `json:"syncState" form:"syncState" `
	PageNum   int                  `json:"pageNum" form:"pageNum"`
	PageSize  int                  `json:"pageSize" form:"pageSize"`
}

type UserListRsp struct {
	Total int          `json:"total"`
	Users []model.User `json:"users"`
}

// List 记录列表
func List(c *gin.Context) {
	var req UserListReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	var users = model.NewUsers()
	err = users.List(
		&model.User{
			Username:  req.Username,
			Nickname:  req.Nickname,
			Status:    req.Status,
			SyncState: req.SyncState,
		},
		req.PageNum,
		req.PageSize,
	)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取用户列表失败: "+err.Error())))
		return
	}

	rets := make([]model.User, 0)
	for _, user := range users {
		rets = append(rets, *user)
	}
	count, err := model.UserListCount(
		&model.User{
			Username:  req.Username,
			Nickname:  req.Nickname,
			Status:    req.Status,
			SyncState: req.SyncState,
		},
	)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取用户总数失败: "+err.Error())))
		return
	}

	helper.Success(c, UserListRsp{
		Total: int(count),
		Users: rets,
	})
}

// UserDeleteReq 批量删除资源结构体
type UserDeleteReq struct {
	UserIds []uint `json:"userIds" validate:"required"`
}

// Delete 删除记录
func Delete(c *gin.Context) {
	var req UserDeleteReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	for _, id := range req.UserIds {
		filter := map[string]interface{}{"id": int(id)}
		var u model.User
		if !u.Exist(filter) {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("有用户不存在")))
			return
		}
	}

	// 根据用户ID获取用户角色排序最小值
	roleMinSortList, err := model.GetUserMinRoleSortsByIds(req.UserIds)
	if err != nil || len(roleMinSortList) == 0 {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("根据用户ID获取用户角色排序最小值失败")))
		return
	}

	// 获取当前登陆用户角色排序最小值（最高等级角色）以及当前用户
	minSort, ctxUser, err := auth.GetCtxLoginUserMinRole(c)
	if err != nil {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("获取当前登陆用户角色排序最小值失败")))
		return
	}

	// 不能删除自己
	if funk.Contains(req.UserIds, ctxUser.ID) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("用户不能删除自己")))
		return
	}

	// 不能删除比自己(登陆用户)角色排序低(等级高)的用户
	for _, sort := range roleMinSortList {
		if int(minSort) > sort {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("用户不能删除比自己角色等级高的用户")))
			return
		}
	}

	var users = model.NewUsers()
	err = users.GetUsersByIds(req.UserIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取用户信息失败: "+err.Error())))
		return
	}

	// 先将用户从ldap中删除
	if config.Conf.Ldap.EnableManage {
		for _, user := range users {
			err := ldapmgr.LdapUserDelete(user.UserDN)
			if err != nil {
				helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("在LDAP删除用户失败"+err.Error())))
				return
			}
		}
	}

	// 再将用户从MySQL中删除
	err = model.DeleteUsersById(req.UserIds)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("在MySQL删除用户失败: "+err.Error())))
		return
	}

	// 删除缓存
	model.CacheUserInfoClear()

	if config.Conf.Notice.DefaultNoticeSwitch {
		// Notifications to users by role's keyword
		keyword := config.Conf.Notice.DefaultNoticeRoleKeyword
		noticeUsers, err := model.GetRoleUsersByKeyword(keyword)
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("通知 %s 组失败, 获取主邮箱失败: %v", keyword, err.Error())))
			return
		}
		noticeUsersEmail := []string{}
		for _, user := range noticeUsers {
			noticeUsersEmail = append(noticeUsersEmail, user.Mail)
		}

		delUsernames := []string{}
		for _, user := range users {
			delUsernames = append(delUsernames, user.Username)
		}
		if err := tools.SendUserStatusNotifications(noticeUsersEmail, delUsernames, "deleted"); err != nil {
			helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("邮件发送删除用户通知失败, 请手工通知"+err.Error())))
			return
		}
	}
	helper.Success(c, nil)
}

// UserResetTotpSecret 重置 Totp 秘钥请求结构体
type UserResetTotpSecret struct {
	Totp string `json:"totp" validate:"required,number,len=6"`
}

// ReSetTotpSecret 重置 Totp 秘钥
func ReSetTotpSecret(c *gin.Context) {
	var req UserResetTotpSecret
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}
	// 获取当前用户
	user, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败")))
		return
	}

	if !totpModel.CheckTotp(user.Totp.Secret, req.Totp) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("OTP验证失败, 如原 TOTP 秘钥无法找回, 可以联系管理员 Update 您的用户信息重新获取")))
		return
	}

	user.Totp.ReSetTotpSecret()
	qrCodeStr := user.GetQrcodestr()

	// 删除缓存
	model.CacheUserInfoDel(user.Username)

	helper.Success(c, qrCodeStr)
}

// UserChangePwdReq 修改密码结构体
type UserChangePwdReq struct {
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
}

// ChangePwd 使用原密码修改密码
func ChangePwd(c *gin.Context) {
	var req UserChangePwdReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 前端传来的密码是rsa加密的,先解密
	// 密码通过RSA解密
	decodeOldPassword, err := tools.RSADecrypt([]byte(req.OldPassword), config.Conf.System.RSAPrivateBytes)
	if err != nil {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("原密码解析失败")))
		return
	}
	decodeNewPassword, err := tools.RSADecrypt([]byte(req.NewPassword), config.Conf.System.RSAPrivateBytes)
	if err != nil {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("新密码解析失败")))
		return
	}
	req.OldPassword = string(decodeOldPassword)
	req.NewPassword = string(decodeNewPassword)

	if complex := tools.CheckPasswordComplexity(req.NewPassword); !complex {
		helper.ErrV2(c, helper.NewValidatorError(tools.ErrPasswordNotComplex))
		return
	}

	// 获取当前用户
	user, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败")))
		return
	}
	// 获取用户的真实正确密码
	// correctPasswd := user.Password
	// 判断前端请求的密码是否等于真实密码
	// err = helper.ComparePasswd(correctPasswd, r.OldPassword)
	// if err != nil {
	// 	helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("原密码错误")))
	// 	return
	// }
	if tools.NewParsePasswd(user.Password) != req.OldPassword {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("原密码错误")))
		return
	}

	// ldap更新密码时可以直接指定用户DN和新密码即可更改成功
	if config.Conf.Ldap.EnableManage {
		err = ldapmgr.LdapUserChangePwd(user.UserDN, "", req.NewPassword)
		if err != nil {
			helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("在LDAP更新密码失败"+err.Error())))
			return
		}
	}

	// 更新密码
	err = user.ChangePwd(tools.NewGenPasswd(req.NewPassword))
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("在MySQL更新密码失败: "+err.Error())))
		return
	}

	// 删除缓存
	model.CacheUserInfoDel(user.Username)

	helper.Success(c, nil)
}

// UserChangeUserStatusReq 修改用户状态结构体
type UserChangeUserStatusReq struct {
	ID     uint             `json:"id" validate:"required"`
	Status model.UserStatus `json:"status" validate:"oneof=1 2"`
}

// ChangeUserStatus 修改用户状态
func ChangeUserStatus(c *gin.Context) {
	var req UserChangeUserStatusReq
	err := helper.BindAndValidateRequest(c, &req)
	if err != nil {
		return
	}

	// 校验工作
	filter := map[string]interface{}{"id": req.ID}
	var u model.User
	if !u.Exist(filter) {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("该用户不存在")))
		return
	}
	user := new(model.User)
	err = user.Find(filter)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("在MySQL查询用户失败: "+err.Error())))
		return
	}
	if req.Status == user.Status {
		if req.Status == 2 {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("用户已经是禁用状态")))
			return
		} else if req.Status == 1 {
			helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("用户已经是启用状态")))
			return
		}
	}
	// 获取当前登录用户，只有管理员才能够将用户状态改变
	// 获取当前登陆用户角色排序最小值（最高等级角色）以及当前用户
	minSort, _, err := auth.GetCtxLoginUserMinRole(c)
	if err != nil {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("获取当前登陆用户角色排序最小值失败")))
		return
	}

	if int(minSort) != 1 {
		helper.ErrV2(c, helper.NewValidatorError(fmt.Errorf("只有管理员才能更改用户状态")))
		return
	}

	var statusDesc string
	var syncStat model.UserSyncStatus
	if req.Status == model.UserDisabled {
		if config.Conf.Ldap.EnableManage {
			err = ldapmgr.LdapUserDelete(user.UserDN)
			if err != nil {
				helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("在LDAP删除用户失败"+err.Error())))
				return
			}
		}
		statusDesc = "deactivated"
		syncStat = model.UserSyncUnNormal
	} else if req.Status == model.UserNormal {
		if config.Conf.Ldap.EnableManage {
			err = ldapmgr.LdapUserAdd(user)
			if err != nil {
				helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("在LDAP添加用户失败"+err.Error())))
				return
			}
		}
		statusDesc = "actived"
		syncStat = model.UserSyncNormal
	}

	err = user.ChangeStatus(req.Status, syncStat)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("在MySQL更新用户状态失败: "+err.Error())))
		return
	}

	// 删除缓存
	model.CacheUserInfoDel(user.Username)

	if config.Conf.Notice.DefaultNoticeSwitch {
		// Notifications to users by role's keyword
		keyword := config.Conf.Notice.DefaultNoticeRoleKeyword
		noticeUsers, err := model.GetRoleUsersByKeyword(keyword)
		if err != nil {
			helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("通知 %s 组失败, 获取主邮箱失败: %v", keyword, err.Error())))
			return
		}
		noticeUsersEmail := []string{}
		for _, user := range noticeUsers {
			noticeUsersEmail = append(noticeUsersEmail, user.Mail)
		}

		usernames := []string{user.Username}
		if err := tools.SendUserStatusNotifications(noticeUsersEmail, usernames, statusDesc); err != nil {
			helper.ErrV2(c, helper.NewLdapError(fmt.Errorf("邮件发送变更用户通知失败, 请手工通知"+err.Error())))
			return
		}
	}

	helper.Success(c, nil)
}

// GetUserInfo 获取当前登录用户信息
func GetUserInfo(c *gin.Context) {
	user, err := auth.GetCtxLoginUser(c)
	if err != nil {
		helper.ErrV2(c, helper.NewMySqlError(fmt.Errorf("获取当前用户信息失败: "+err.Error())))
		return
	}
	helper.Success(c, user)
}
