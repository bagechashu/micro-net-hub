package user

import (
	"errors"
	"fmt"
	"strings"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/module/goldap/ldapmgr"
	totpModel "micro-net-hub/internal/module/totp/model"
	userModel "micro-net-hub/internal/module/user/model"
	"micro-net-hub/internal/server/helper"
	"micro-net-hub/internal/tools"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/thoas/go-funk"
)

type UserLogic struct{}

// Add 添加数据
func (l UserLogic) Add(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.UserAddReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	var u userModel.User
	if u.Exist(tools.H{"username": r.Username}) {
		return nil, helper.NewValidatorError(fmt.Errorf("用户名已存在,请勿重复添加"))
	}
	// if u.Exist(tools.H{"mobile": r.Mobile}) {
	// 	return nil, helper.NewValidatorError(fmt.Errorf("手机号已存在,请勿重复添加"))
	// }
	// if u.Exist(tools.H{"job_number": r.JobNumber}) {
	// 	return nil, helper.NewValidatorError(fmt.Errorf("工号已存在,请勿重复添加"))
	// }
	if u.Exist(tools.H{"mail": r.Mail}) {
		return nil, helper.NewValidatorError(fmt.Errorf("邮箱已存在,请勿重复添加"))
	}

	// 密码通过RSA解密
	// 密码不为空就解密
	if r.Password != "" {
		decodeData, err := tools.RSADecrypt([]byte(r.Password), config.Conf.System.RSAPrivateBytes)
		if err != nil {
			return nil, helper.NewValidatorError(fmt.Errorf("密码解密失败"))
		}
		r.Password = string(decodeData)
		if len(r.Password) < 6 {
			return nil, helper.NewValidatorError(fmt.Errorf("密码长度至少为6位"))
		}
	} else {
		r.Password = tools.GeneratePassword(8)
	}

	// 当前登陆用户角色排序最小值（最高等级角色）以及当前登陆的用户
	currentRoleSortMin, ctxUser, err := GetCurrentUserMinRoleSort(c)
	if err != nil {
		return nil, helper.NewValidatorError(fmt.Errorf("获取当前登陆用户角色排序最小值失败"))
	}

	// 根据角色id获取角色
	if r.RoleIds == nil || len(r.RoleIds) == 0 {
		r.RoleIds = []uint{2} // 默认添加为普通用户角色
	}

	roles := userModel.NewRoles()
	err = roles.GetRolesByIds(r.RoleIds)
	if err != nil {
		return nil, helper.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败"))
	}

	var reqRoleSorts []int
	for _, role := range roles {
		reqRoleSorts = append(reqRoleSorts, int(role.Sort))
	}
	// 前端传来用户角色排序最小值（最高等级角色）
	reqRoleSortMin := uint(funk.MinInt(reqRoleSorts).(int))

	// 如果登录用户的角色ID为1，亦即为管理员，则直接放行，保障管理员拥有最大权限
	if currentRoleSortMin != 1 {
		// 当前用户的角色排序最小值 需要小于 前端传来的角色排序最小值（用户不能创建比自己等级高的或者相同等级的用户）
		if currentRoleSortMin >= reqRoleSortMin {
			return nil, helper.NewValidatorError(fmt.Errorf("用户不能创建比自己等级高的或者相同等级的用户"))
		}
	}
	user := userModel.User{
		Username:      r.Username,
		Password:      r.Password,
		Nickname:      r.Nickname,
		GivenName:     r.GivenName,
		Mail:          r.Mail,
		JobNumber:     r.JobNumber,
		Mobile:        r.Mobile,
		Avatar:        r.Avatar,
		PostalAddress: r.PostalAddress,
		Departments:   r.Departments,
		Position:      r.Position,
		Introduction:  r.Introduction,
		Status:        r.Status,
		Creator:       ctxUser.Username,
		DepartmentId:  tools.SliceToString(r.DepartmentId, ","),
		Source:        r.Source,
		Roles:         roles,
		UserDN:        fmt.Sprintf("uid=%s,%s", r.Username, config.Conf.Ldap.UserDN),
	}

	if user.Source == "" {
		user.Source = "platform"
	}

	// 获取用户将要添加的分组

	var gs = userModel.NewGroups()
	err = gs.GetGroupsByIds(tools.StringToSlice(user.DepartmentId, ","))
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("根据部门ID获取部门信息失败: " + err.Error()))
	}

	err = CommonAddUser(&user, gs)
	if err != nil {
		return nil, helper.NewOperationError(fmt.Errorf("添加用户失败: " + err.Error()))
	}

	if r.Notice {
		var nu userModel.User
		if nu.Find(tools.H{"username": r.Username}) != nil {
			return nil, helper.NewValidatorError(fmt.Errorf("系统通知用户账号信息失败, 请手工通知"))
		}
		qrRawPngBase64, err := nu.GetRawPngBase64()
		if err != nil {
			return nil, helper.NewOperationError(fmt.Errorf("系统通知用户账号信息失败, 请手工通知"))
		}
		tools.SendUserInfo([]string{nu.Mail}, nu.Username, r.Password, qrRawPngBase64)
	}

	return nil, nil
}

// List 数据列表
func (l UserLogic) List(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.UserListReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	var users = userModel.NewUsers()
	err := users.List(r)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取用户列表失败: " + err.Error()))
	}

	rets := make([]userModel.User, 0)
	for _, user := range users {
		rets = append(rets, *user)
	}
	count, err := userModel.UserListCount(r)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取用户总数失败: " + err.Error()))
	}

	return userModel.UserListRsp{
		Total: int(count),
		Users: rets,
	}, nil
}

// Update 更新数据
func (l UserLogic) Update(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.UserUpdateReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	var u userModel.User
	if !u.Exist(tools.H{"id": r.ID}) {
		return nil, helper.NewMySqlError(fmt.Errorf("该记录不存在"))
	}

	// 获取当前登陆用户
	ctxUser, err := GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败"))
	}

	// 获取当前登陆用户角色ID集合
	var currentRoleSorts []int
	for _, role := range ctxUser.Roles {
		currentRoleSorts = append(currentRoleSorts, int(role.Sort))
	}

	// 获取将要操作的用户角色ID集合
	var reqRoleSorts []int

	roles := userModel.NewRoles()
	err = roles.GetRolesByIds(r.RoleIds)
	if err != nil {
		return nil, helper.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败"))
	}
	if len(roles) == 0 {
		return nil, helper.NewValidatorError(fmt.Errorf("根据角色ID获取角色信息失败"))
	}
	for _, role := range roles {
		reqRoleSorts = append(reqRoleSorts, int(role.Sort))
	}

	// 当前登陆用户角色排序最小值（最高等级角色）
	currentRoleSortMin := funk.MinInt(currentRoleSorts).(int)
	// 前端传来用户角色排序最小值（最高等级角色）
	reqRoleSortMin := funk.MinInt(reqRoleSorts).(int)

	// 如果登录用户的角色ID为1，亦即为管理员，则直接放行，保障管理员拥有最大权限
	if currentRoleSortMin != 1 {
		// 判断是更新自己还是更新别人,如果操作的ID与登陆用户的ID一致，则说明操作的是自己
		if int(r.ID) == int(ctxUser.ID) {
			// 不能更改自己的角色
			reqDiff, currentDiff := funk.Difference(reqRoleSorts, currentRoleSorts)
			if len(reqDiff.([]int)) > 0 || len(currentDiff.([]int)) > 0 {
				return nil, helper.NewValidatorError(fmt.Errorf("不能更改自己的角色"))
			}
		}

		// 如果是更新别人，操作者不能更新比自己角色等级高的或者相同等级的用户
		minRoleSorts, err := userModel.GetUserMinRoleSortsByIds([]uint{uint(r.ID)}) // 根据userIdID获取用户角色排序最小值
		if err != nil || len(minRoleSorts) == 0 {
			return nil, helper.NewValidatorError(fmt.Errorf("根据用户ID获取用户角色排序最小值失败"))
		}
		if currentRoleSortMin >= minRoleSorts[0] || currentRoleSortMin >= reqRoleSortMin {
			return nil, helper.NewValidatorError(fmt.Errorf("用户不能更新比自己角色等级高的或者相同等级的用户"))
		}
	}

	// 先获取用户信息
	oldData := new(userModel.User)
	err = oldData.Find(tools.H{"id": r.ID})
	if err != nil {
		return nil, helper.NewMySqlError(err)
	}

	// 过滤掉前端会选择到的 请选择部门信息 这个选项
	var (
		depts   string
		deptids []uint
	)
	for _, v := range strings.Split(r.Departments, ",") {
		if v != "请选择部门信息" {
			depts += v + ","
		}
	}
	for _, j := range r.DepartmentId {
		if j != 0 {
			deptids = append(deptids, j)
		}
	}

	// 拼装新的用户信息
	user := userModel.User{
		Model:         oldData.Model,
		Username:      r.Username,
		Nickname:      r.Nickname,
		GivenName:     r.GivenName,
		Mail:          r.Mail,
		JobNumber:     r.JobNumber,
		Mobile:        r.Mobile,
		Avatar:        r.Avatar,
		PostalAddress: r.PostalAddress,
		Departments:   depts,
		Position:      r.Position,
		Introduction:  r.Introduction,
		Creator:       ctxUser.Username,
		DepartmentId:  tools.SliceToString(deptids, ","),
		Source:        oldData.Source,
		Roles:         roles,
		UserDN:        oldData.UserDN,
	}

	// 密码不为空就解密并更新, 为空则不更新
	if r.Password != "" {
		decodeData, err := tools.RSADecrypt([]byte(r.Password), config.Conf.System.RSAPrivateBytes)
		if err != nil {
			return nil, helper.NewValidatorError(fmt.Errorf("密码解密失败"))
		}
		r.Password = string(decodeData)
		if len(r.Password) < 6 {
			return nil, helper.NewValidatorError(fmt.Errorf("密码长度至少为6位"))
		}
		user.Password = r.Password
	} else {
		r.Password = "The password has not been updated. Please continue using your current password."
	}

	if err = CommonUpdateUser(oldData, &user, r.DepartmentId); err != nil {
		return nil, helper.NewOperationError(fmt.Errorf("更新用户失败: " + err.Error()))
	}

	// flush user info cache
	userModel.ClearUserInfoCache()

	if r.Notice {
		var nu userModel.User
		if nu.Find(tools.H{"username": r.Username}) != nil {
			return nil, helper.NewValidatorError(fmt.Errorf("系统通知用户账号信息失败, 请手工通知"))
		}
		qrRawPngBase64, err := nu.GetRawPngBase64()
		if err != nil {
			return nil, helper.NewOperationError(fmt.Errorf("系统通知用户账号信息失败, 请手工通知"))
		}
		tools.SendUserInfo([]string{nu.Mail}, nu.Username, r.Password, qrRawPngBase64)
	}

	return nil, nil
}

// Delete 删除数据
func (l UserLogic) Delete(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.UserDeleteReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c

	for _, id := range r.UserIds {
		filter := tools.H{"id": int(id)}
		var u userModel.User
		if !u.Exist(filter) {
			return nil, helper.NewMySqlError(fmt.Errorf("有用户不存在"))
		}
	}

	// 根据用户ID获取用户角色排序最小值
	roleMinSortList, err := userModel.GetUserMinRoleSortsByIds(r.UserIds)
	if err != nil || len(roleMinSortList) == 0 {
		return nil, helper.NewValidatorError(fmt.Errorf("根据用户ID获取用户角色排序最小值失败"))
	}

	// 获取当前登陆用户角色排序最小值（最高等级角色）以及当前用户
	minSort, ctxUser, err := GetCurrentUserMinRoleSort(c)
	if err != nil {
		return nil, helper.NewValidatorError(fmt.Errorf("获取当前登陆用户角色排序最小值失败"))
	}

	// 不能删除自己
	if funk.Contains(r.UserIds, ctxUser.ID) {
		return nil, helper.NewValidatorError(fmt.Errorf("用户不能删除自己"))
	}

	// 不能删除比自己(登陆用户)角色排序低(等级高)的用户
	for _, sort := range roleMinSortList {
		if int(minSort) > sort {
			return nil, helper.NewValidatorError(fmt.Errorf("用户不能删除比自己角色等级高的用户"))
		}
	}

	var users = userModel.NewUsers()
	err = users.GetUserByIds(r.UserIds)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取用户信息失败: " + err.Error()))
	}

	// 先将用户从ldap中删除
	for _, user := range users {
		err := ldapmgr.LdapUserDelete(user.UserDN)
		if err != nil {
			return nil, helper.NewLdapError(fmt.Errorf("在LDAP删除用户失败" + err.Error()))
		}
	}

	// 再将用户从MySQL中删除
	err = userModel.DeleteUsersById(r.UserIds)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("在MySQL删除用户失败: " + err.Error()))
	}

	// flush user info cache
	userModel.ClearUserInfoCache()

	if config.Conf.Notice.DefaultNoticeSwitch {
		// Notifications to users by role's keyword
		keyword := config.Conf.Notice.DefaultNoticeRoleKeyword
		noticeUsers, err := userModel.GetRoleUsersByKeyword(keyword)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("通知 %s 组失败, 获取主邮箱失败: %v", keyword, err.Error()))
		}
		noticeUsersEmail := []string{}
		for _, user := range noticeUsers {
			noticeUsersEmail = append(noticeUsersEmail, user.Mail)
		}

		delUsernames := []string{}
		for _, user := range users {
			delUsernames = append(delUsernames, user.Username)
		}
		tools.SendUserStatusNotifications(noticeUsersEmail, delUsernames, "deleted")
	}

	return nil, nil
}

// Reset TOTP Secret 重置 TOTP 密钥
func (l UserLogic) ReSetTotpSecret(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.UserResetTotpSecret)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	// 获取当前用户
	user, err := GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败"))
	}

	if !totpModel.CheckTotp(user.Totp.Secret, r.Totp) {
		return nil, helper.NewValidatorError(fmt.Errorf("OTP验证失败, 如原 TOTP 秘钥无法找回, 可以联系管理员 Update 您的用户信息重新获取"))
	}

	user.Totp.ReSetTotpSecret()
	qrCodeStr := user.GetQrcodestr()

	return qrCodeStr, nil
}

// ChangePwd 修改密码
func (l UserLogic) ChangePwd(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.UserChangePwdReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c
	// 前端传来的密码是rsa加密的,先解密
	// 密码通过RSA解密
	decodeOldPassword, err := tools.RSADecrypt([]byte(r.OldPassword), config.Conf.System.RSAPrivateBytes)
	if err != nil {
		return nil, helper.NewValidatorError(fmt.Errorf("原密码解析失败"))
	}
	decodeNewPassword, err := tools.RSADecrypt([]byte(r.NewPassword), config.Conf.System.RSAPrivateBytes)
	if err != nil {
		return nil, helper.NewValidatorError(fmt.Errorf("新密码解析失败"))
	}
	r.OldPassword = string(decodeOldPassword)
	r.NewPassword = string(decodeNewPassword)
	// 获取当前用户
	user, err := GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前登陆用户失败"))
	}
	// 获取用户的真实正确密码
	// correctPasswd := user.Password
	// 判断前端请求的密码是否等于真实密码
	// err = helper.ComparePasswd(correctPasswd, r.OldPassword)
	// if err != nil {
	// 	return nil, helper.NewValidatorError(fmt.Errorf("原密码错误"))
	// }
	if tools.NewParsePasswd(user.Password) != r.OldPassword {
		return nil, helper.NewValidatorError(fmt.Errorf("原密码错误"))
	}
	// ldap更新密码时可以直接指定用户DN和新密码即可更改成功
	err = ldapmgr.LdapUserChangePwd(user.UserDN, "", r.NewPassword)
	if err != nil {
		return nil, helper.NewLdapError(fmt.Errorf("在LDAP更新密码失败" + err.Error()))
	}

	// 更新密码
	err = user.ChangePwd(tools.NewGenPasswd(r.NewPassword))
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("在MySQL更新密码失败: " + err.Error()))
	}

	return nil, nil
}

// ChangeUserStatus 修改用户状态
func (l UserLogic) ChangeUserStatus(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.UserChangeUserStatusReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}
	_ = c
	// 校验工作
	filter := tools.H{"id": r.ID}
	var u userModel.User
	if !u.Exist(filter) {
		return nil, helper.NewValidatorError(fmt.Errorf("该用户不存在"))
	}
	user := new(userModel.User)
	err := user.Find(filter)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("在MySQL查询用户失败: " + err.Error()))
	}
	if r.Status == user.Status {
		if r.Status == 2 {
			return nil, helper.NewValidatorError(fmt.Errorf("用户已经是离职状态"))
		} else if r.Status == 1 {
			return nil, helper.NewValidatorError(fmt.Errorf("用户已经是在职状态"))
		}
	}
	// 获取当前登录用户，只有管理员才能够将用户状态改变
	// 获取当前登陆用户角色排序最小值（最高等级角色）以及当前用户
	minSort, _, err := GetCurrentUserMinRoleSort(c)
	if err != nil {
		return nil, helper.NewValidatorError(fmt.Errorf("获取当前登陆用户角色排序最小值失败"))
	}

	if int(minSort) != 1 {
		return nil, helper.NewValidatorError(fmt.Errorf("只有管理员才能更改用户状态"))
	}

	var statusDesc string
	if r.Status == 2 {
		err = ldapmgr.LdapUserDelete(user.UserDN)
		if err != nil {
			return nil, helper.NewLdapError(fmt.Errorf("在LDAP删除用户失败" + err.Error()))
		}
		statusDesc = "deactivated"
	} else if r.Status == 1 {
		err = ldapmgr.LdapUserAdd(user)
		if err != nil {
			return nil, helper.NewLdapError(fmt.Errorf("在LDAP添加用户失败" + err.Error()))
		}
		statusDesc = "actived"
	}

	err = user.ChangeStatus(int(r.Status))
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("在MySQL更新用户状态失败: " + err.Error()))
	}

	if config.Conf.Notice.DefaultNoticeSwitch {
		// Notifications to users by role's keyword
		keyword := config.Conf.Notice.DefaultNoticeRoleKeyword
		noticeUsers, err := userModel.GetRoleUsersByKeyword(keyword)
		if err != nil {
			return nil, helper.NewMySqlError(fmt.Errorf("通知 %s 组失败, 获取主邮箱失败: %v", keyword, err.Error()))
		}
		noticeUsersEmail := []string{}
		for _, user := range noticeUsers {
			noticeUsersEmail = append(noticeUsersEmail, user.Mail)
		}

		usernames := []string{user.Username}
		tools.SendUserStatusNotifications(noticeUsersEmail, usernames, statusDesc)
	}

	return nil, nil
}

// GetUserInfo 获取用户信息
func (l UserLogic) GetUserInfo(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	r, ok := req.(*userModel.UserInfoReq)
	if !ok {
		return nil, helper.ReqAssertErr
	}

	_ = c
	_ = r

	user, err := GetCurrentLoginUser(c)
	if err != nil {
		return nil, helper.NewMySqlError(fmt.Errorf("获取当前用户信息失败: " + err.Error()))
	}
	return user, nil
}

// GetCurrentUserMinRoleSort  获取当前用户角色排序最小值（最高等级角色）以及当前用户信息
func GetCurrentUserMinRoleSort(c *gin.Context) (uint, userModel.User, error) {
	// 获取当前用户
	ctxUser, err := GetCurrentLoginUser(c)
	if err != nil {
		return 999, ctxUser, err
	}
	// 获取当前用户的所有角色
	currentRoles := ctxUser.Roles
	// 获取当前用户角色的排序，和前端传来的角色排序做比较
	var currentRoleSorts []int
	for _, role := range currentRoles {
		currentRoleSorts = append(currentRoleSorts, int(role.Sort))
	}
	// 当前用户角色排序最小值（最高等级角色）
	currentRoleSortMin := uint(funk.MinInt(currentRoleSorts).(int))

	return currentRoleSortMin, ctxUser, nil
}

// GetCurrentLoginUser 获取当前登录用户信息
// 需要缓存，减少数据库访问
func GetCurrentLoginUser(c *gin.Context) (userModel.User, error) {
	var newUser userModel.User
	ctxUser, exist := c.Get("user")
	if !exist {
		return newUser, errors.New("用户未登录")
	}
	u, _ := ctxUser.(userModel.User)

	// 先获取缓存
	cacheUser, found := userModel.UserInfoCache.Get(u.Username)
	var user userModel.User
	var err error
	if found {
		user = cacheUser.(userModel.User)
		err = nil
	} else {
		// 缓存中没有就获取数据库
		err = user.GetUserById(u.ID)
		// 获取成功就缓存
		if err != nil {
			userModel.UserInfoCache.Delete(u.Username)
		} else {
			userModel.UserInfoCache.Set(u.Username, user, cache.DefaultExpiration)
		}
	}
	return user, err
}
