package setup

import (
	"errors"
	"fmt"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	accountModel "micro-net-hub/internal/module/account/model"
	apiMgrModel "micro-net-hub/internal/module/apimgr/model"
	sitenavModel "micro-net-hub/internal/module/sitenav/model"

	"github.com/thoas/go-funk"
	"gorm.io/gorm"
)

// 初始化mysql数据
func InitData() {
	// 是否初始化数据
	if !config.Conf.System.InitData {
		return
	}

	u := new(accountModel.User)
	err := global.DB.First(u, gorm.Model{ID: 1}).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Log.Warnf("数据库中已存在用户数据，不进行初始化数据操作!")
		return
	}

	// 1.写入角色数据
	roles := []*accountModel.Role{
		{
			Model:   gorm.Model{ID: accountModel.SuperAdminRoleID},
			Name:    "SuperAdmin",
			Keyword: "superadmin",
			Remark:  "",
			Sort:    1,
			Status:  1,
			Creator: "System",
		},
		{
			Model:   gorm.Model{ID: 2},
			Name:    "Users",
			Keyword: "user",
			Remark:  "",
			Sort:    3,
			Status:  1,
			Creator: "System",
		},
		{
			Model:   gorm.Model{ID: 3},
			Name:    "Guests",
			Keyword: "guest",
			Remark:  "",
			Sort:    5,
			Status:  1,
			Creator: "System",
		},
	}

	// newRoles := make([]*accountModel.Role, 0)
	// for _, role := range roles {
	// 	err := global.DB.First(&role, role.ID).Error
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		newRoles = append(newRoles, role)
	// 	}
	// }

	// if len(newRoles) > 0 {
	// 	err := global.DB.Create(&newRoles).Error
	// 	if err != nil {
	// 		global.Log.Errorf("写入系统角色数据失败：%s", err)
	// 	} else {
	// 		global.Log.Infof("写入系统角色数据成功: %v", newRoles)
	// 	}
	// }

	err = global.DB.Create(&roles).Error
	if err != nil {
		global.Log.Errorf("写入系统角色数据失败：%s", err)
	} else {
		global.Log.Info("写入系统角色数据成功")
	}

	// 2写入菜单
	var basicGID uint = 0
	var userMgrGID uint = 1
	var sysMgrGID uint = 2
	var OperationLogGID uint = 3

	menus := []accountModel.Menu{
		{
			Model:     gorm.Model{ID: userMgrGID},
			Name:      "UserManage",
			Title:     "人员管理",
			Icon:      "user",
			Path:      "/personnel",
			Component: "Layout",
			Redirect:  "/personnel/user",
			Sort:      1,
			ParentId:  basicGID,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Model:     gorm.Model{ID: sysMgrGID},
			Name:      "System",
			Title:     "系统管理",
			Icon:      "component",
			Path:      "/system",
			Component: "Layout",
			Redirect:  "/system/role",
			Sort:      2,
			ParentId:  basicGID,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Model:     gorm.Model{ID: OperationLogGID},
			Name:      "Log",
			Title:     "日志管理",
			Icon:      "example",
			Path:      "/log",
			Component: "Layout",
			Redirect:  "/log/operation-log",
			Sort:      3,
			ParentId:  basicGID,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:      "Profile",
			Title:     "个人中心",
			Icon:      "people",
			Path:      "/profile/index",
			Component: "Layout",
			Sort:      4,
			ParentId:  basicGID,
			Roles:     roles[:2],
			Creator:   "System",
		},
		{
			Name:      "User",
			Title:     "用户管理",
			Icon:      "people",
			Path:      "user",
			Component: "/personnel/user/index",
			Sort:      11,
			ParentId:  userMgrGID,
			NoCache:   1,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:      "Group",
			Title:     "分组管理",
			Icon:      "peoples",
			Path:      "group",
			Component: "/personnel/group/index",
			Sort:      12,
			ParentId:  userMgrGID,
			NoCache:   1,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:      "FieldRelation",
			Title:     "字段关系管理",
			Icon:      "el-icon-s-tools",
			Path:      "fieldRelation",
			Component: "/personnel/fieldRelation/index",
			Sort:      13,
			ParentId:  userMgrGID,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:      "Role",
			Title:     "角色管理",
			Icon:      "eye-open",
			Path:      "role",
			Component: "/system/role/index",
			Sort:      21,
			ParentId:  sysMgrGID,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:      "Menu",
			Title:     "菜单管理",
			Icon:      "tree-table",
			Path:      "menu",
			Component: "/system/menu/index",
			Sort:      22,
			ParentId:  sysMgrGID,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:      "Api",
			Title:     "接口管理",
			Icon:      "tree",
			Path:      "api",
			Component: "/system/api/index",
			Sort:      23,
			ParentId:  sysMgrGID,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:      "Sitenav",
			Title:     "导航配置",
			Icon:      "list",
			Path:      "sitenavmgr",
			Component: "/sitenav/manager",
			Sort:      24,
			ParentId:  sysMgrGID,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:      "DnsManager",
			Title:     "Dns配置",
			Icon:      "guide",
			Path:      "dnsmgr",
			Component: "/dns/manager",
			Sort:      25,
			ParentId:  sysMgrGID,
			NoCache:   1,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:      "NoticeManager",
			Title:     "公告管理",
			Icon:      "message",
			Path:      "noticemgr",
			Component: "/notice/manager",
			Sort:      26,
			ParentId:  sysMgrGID,
			NoCache:   1,
			Roles:     roles[:1],
			Creator:   "System",
		},
		{
			Name:      "OperationLog",
			Title:     "操作日志",
			Icon:      "documentation",
			Path:      "operation-log",
			Component: "/log/operation-log/index",
			Sort:      31,
			ParentId:  OperationLogGID,
			Roles:     roles[:1],
			Creator:   "System",
		},
	}

	// newMenus := make([]accountModel.Menu, 0)
	// for _, menu := range menus {
	// 	err := global.DB.First(&menu, menu.ID).Error
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		newMenus = append(newMenus, menu)
	// 	}
	// }
	// if len(newMenus) > 0 {
	// 	err := global.DB.Create(&newMenus).Error
	// 	if err != nil {
	// 		global.Log.Errorf("写入系统菜单数据失败：%s", err)
	// 	} else {
	// 		global.Log.Infof("写入系统菜单数据成功: %v", newMenus)
	// 	}
	// }

	err = global.DB.Create(&menus).Error
	if err != nil {
		global.Log.Errorf("写入系统菜单数据失败：%s", err)
	} else {
		global.Log.Info("写入系统菜单数据成功")
	}

	// 3.写入用户
	users := []*accountModel.User{
		{
			Model:         gorm.Model{ID: 1},
			Username:      "admin",
			Password:      config.Conf.Ldap.AdminPass,
			Nickname:      "Super Admin",
			GivenName:     "Super Admin",
			Mail:          "admin@" + config.Conf.Ldap.DefaultEmailSuffix,
			JobNumber:     "0000",
			Mobile:        "18888888888",
			Avatar:        "https://q1.qlogo.cn/g?b=qq&nk=10002&s=100",
			PostalAddress: "default",
			Position:      "default",
			Introduction:  "default",
			Status:        1,
			Creator:       "System",
			Roles:         roles[:1],
			UserDN:        config.Conf.Ldap.AdminDN,
		},
	}

	// newUsers := make([]*accountModel.User, 0)
	// for _, user := range users {
	// 	err := global.DB.First(&user, user.ID).Error
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		newUsers = append(newUsers, user)
	// 	}
	// }

	// if len(newUsers) > 0 {
	// 	err := global.DB.Create(&newUsers).Error
	// 	if err != nil {
	// 		global.Log.Errorf("写入用户数据失败：%s", err)
	// 	} else {
	// 		global.Log.Infof("写入用户数据成功: %v", newUsers)
	// 	}
	// }

	err = global.DB.Create(&users).Error
	if err != nil {
		global.Log.Errorf("写入用户数据失败：%s", err)
	} else {
		global.Log.Info("写入用户数据成功")
	}

	// 4.写入api
	apis := []apiMgrModel.Api{
		{
			Method:   "POST",
			Path:     "/base/login",
			Category: "base",
			Remark:   "用户登录",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/base/logout",
			Category: "base",
			Remark:   "用户登出",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/base/refreshToken",
			Category: "base",
			Remark:   "刷新JWT令牌",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/base/sendcode",
			Category: "base",
			Remark:   "给用户邮箱发送验证码",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/base/changePwd",
			Category: "base",
			Remark:   "通过邮箱修改密码",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/user/info",
			Category: "user",
			Remark:   "获取当前登录用户信息",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/user/list",
			Category: "user",
			Remark:   "获取用户列表",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/user/resetTotpSecret",
			Category: "user",
			Remark:   "重置用户 TOTP 秘钥",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/user/changePwd",
			Category: "user",
			Remark:   "更新用户登录密码",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/user/add",
			Category: "user",
			Remark:   "创建用户",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/user/update",
			Category: "user",
			Remark:   "更新用户",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/user/delete",
			Category: "user",
			Remark:   "批量删除用户",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/user/changeUserStatus",
			Category: "user",
			Remark:   "更改用户在职状态",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/sync/syncDingTalkUsers",
			Category: "user",
			Remark:   "从钉钉拉取用户信息",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/sync/syncWeComUsers",
			Category: "user",
			Remark:   "从企业微信拉取用户信息",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/sync/syncFeiShuUsers",
			Category: "user",
			Remark:   "从飞书拉取用户信息",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/sync/syncOpenLdapUsers",
			Category: "user",
			Remark:   "从openldap拉取用户信息",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/sync/syncSqlUsers",
			Category: "user",
			Remark:   "将数据库中的用户同步到Ldap",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/group/list",
			Category: "group",
			Remark:   "获取分组列表",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/group/tree",
			Category: "group",
			Remark:   "获取分组列表树",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/group/add",
			Category: "group",
			Remark:   "创建分组",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/group/update",
			Category: "group",
			Remark:   "更新分组",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/group/delete",
			Category: "group",
			Remark:   "批量删除分组",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/group/adduser",
			Category: "group",
			Remark:   "添加用户到分组",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/group/removeuser",
			Category: "group",
			Remark:   "将用户从分组移出",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/group/useringroup",
			Category: "group",
			Remark:   "获取在分组内的用户列表",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/group/usernoingroup",
			Category: "group",
			Remark:   "获取不在分组内的用户列表",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/sync/syncDingTalkDepts",
			Category: "group",
			Remark:   "从钉钉拉取部门信息",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/sync/syncWeComDepts",
			Category: "group",
			Remark:   "从企业微信拉取部门信息",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/sync/syncFeiShuDepts",
			Category: "group",
			Remark:   "从飞书拉取部门信息",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/sync/syncOpenLdapDepts",
			Category: "group",
			Remark:   "从openldap拉取部门信息",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/sync/syncSqlGroups",
			Category: "group",
			Remark:   "将数据库中的分组同步到Ldap",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/role/list",
			Category: "role",
			Remark:   "获取角色列表",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/role/add",
			Category: "role",
			Remark:   "创建角色",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/role/update",
			Category: "role",
			Remark:   "更新角色",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/role/getmenulist",
			Category: "role",
			Remark:   "获取角色的权限菜单",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/role/updatemenus",
			Category: "role",
			Remark:   "更新角色的权限菜单",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/role/getapilist",
			Category: "role",
			Remark:   "获取角色的权限接口",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/role/updateapis",
			Category: "role",
			Remark:   "更新角色的权限接口",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/role/delete",
			Category: "role",
			Remark:   "批量删除角色",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/menu/tree",
			Category: "menu",
			Remark:   "获取菜单树",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/menu/access/tree",
			Category: "menu",
			Remark:   "获取用户菜单树",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/menu/add",
			Category: "menu",
			Remark:   "创建菜单",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/menu/update",
			Category: "menu",
			Remark:   "更新菜单",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/menu/delete",
			Category: "menu",
			Remark:   "批量删除菜单",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/api/list",
			Category: "api",
			Remark:   "获取接口列表",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/api/tree",
			Category: "api",
			Remark:   "获取接口树",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/api/add",
			Category: "api",
			Remark:   "创建接口",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/api/update",
			Category: "api",
			Remark:   "更新接口",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/api/delete",
			Category: "api",
			Remark:   "批量删除接口",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/goldap/fieldrelation/list",
			Category: "fieldrelation",
			Remark:   "获取字段动态关系列表",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/fieldrelation/add",
			Category: "fieldrelation",
			Remark:   "创建字段动态关系",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/fieldrelation/update",
			Category: "fieldrelation",
			Remark:   "更新字段动态关系",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/goldap/fieldrelation/delete",
			Category: "fieldrelation",
			Remark:   "批量删除字段动态关系",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/log/operation/list",
			Category: "log",
			Remark:   "获取操作日志列表",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/log/operation/delete",
			Category: "log",
			Remark:   "批量删除操作日志",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/sitenav/list",
			Category: "sitenav",
			Remark:   "导航配置获取列表",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/sitenav/group/add",
			Category: "sitenav",
			Remark:   "网址导航-组增加",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/sitenav/group/update",
			Category: "sitenav",
			Remark:   "网址导航-组更新",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/sitenav/group/delete",
			Category: "sitenav",
			Remark:   "网址导航-组删除",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/sitenav/site/add",
			Category: "sitenav",
			Remark:   "网址导航-站点增加",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/sitenav/site/update",
			Category: "sitenav",
			Remark:   "网址导航-站点更新",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/sitenav/site/delete",
			Category: "sitenav",
			Remark:   "网址导航-站点删除",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/dns/getall",
			Category: "sitenav",
			Remark:   "网址导航-站点删除",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/dns/zone/add",
			Category: "sitenav",
			Remark:   "网址导航-站点删除",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/dns/zone/update",
			Category: "sitenav",
			Remark:   "网址导航-站点删除",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/dns/zone/delete",
			Category: "sitenav",
			Remark:   "网址导航-站点删除",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/dns/record/add",
			Category: "sitenav",
			Remark:   "网址导航-站点删除",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/dns/record/update",
			Category: "sitenav",
			Remark:   "网址导航-站点删除",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/dns/record/delete",
			Category: "sitenav",
			Remark:   "网址导航-站点删除",
			Creator:  "System",
		},
		{
			Method:   "GET",
			Path:     "/notice/getall",
			Category: "noticeboard",
			Remark:   "首页公告-获取所有公告",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/notice/mgr/add",
			Category: "noticeboard",
			Remark:   "首页公告-增加公告",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/notice/mgr/update",
			Category: "noticeboard",
			Remark:   "首页公告-更新公告",
			Creator:  "System",
		},
		{
			Method:   "POST",
			Path:     "/notice/mgr/delete",
			Category: "noticeboard",
			Remark:   "首页公告-删除公告",
			Creator:  "System",
		},
	}
	if err := global.DB.Create(&apis).Error; err != nil {
		global.Log.Errorf("写入api数据失败：%s", err)
	} else {
		global.Log.Info("写入api数据成功")
	}

	// 5. 将角色绑定给菜单
	// newApi := make([]apiMgrModel.Api, 0)
	// newRoleCasbin := make([]accountModel.RoleCasbin, 0)
	// for i, api := range apis {
	// 	api.ID = uint(i + 1)
	// 	err := global.DB.First(&api, api.ID).Error
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		newApi = append(newApi, api)

	// 		// 管理员拥有所有API权限
	// 		newRoleCasbin = append(newRoleCasbin, accountModel.RoleCasbin{
	// 			Keyword: roles[adminRID].Keyword,
	// 			Path:    api.Path,
	// 			Method:  api.Method,
	// 		})

	// 		// 非管理员拥有基础权限
	// 		basePaths := []string{
	// 			"/base/login",
	// 			"/base/logout",
	// 			"/base/refreshToken",
	// 			"/base/sendcode",
	// 			"/base/changePwd",
	// 			"/base/dashboard",
	// 			"/user/info",
	// 			"/user/resetTotpSecret",
	// 			"/user/changePwd",
	// 			"/menu/access/tree",
	// 			"/sitenav/list",
	// 		}

	// 		if funk.ContainsString(basePaths, api.Path) {
	// 			newRoleCasbin = append(newRoleCasbin, accountModel.RoleCasbin{
	// 				Keyword: roles[userRID].Keyword,
	// 				Path:    api.Path,
	// 				Method:  api.Method,
	// 			})
	// 		}
	// 	}
	// }

	// if len(newApi) > 0 {
	// 	if err := global.DB.Create(&newApi).Error; err != nil {
	// 		global.Log.Errorf("写入api数据失败：%s", err)
	// 	} else {
	// 		global.Log.Infof("写入api数据成功: %v", newApi)
	// 	}
	// }

	// if len(newRoleCasbin) > 0 {
	// 	rules := make([][]string, 0)
	// 	for _, c := range newRoleCasbin {
	// 		rules = append(rules, []string{
	// 			c.Keyword, c.Path, c.Method,
	// 		})
	// 	}
	// 	isAdd, err := global.CasbinEnforcer.AddPolicies(rules)
	// 	if !isAdd {
	// 		global.Log.Errorf("写入casbin数据失败：%v", err)
	// 	}
	// }

	// 管理员拥有所有API权限
	roleCasbin := make([]accountModel.RoleCasbin, 0)
	for _, api := range apis {
		roleCasbin = append(roleCasbin, accountModel.RoleCasbin{
			Keyword: roles[0].Keyword,
			Path:    api.Path,
			Method:  api.Method,
		})

		// 非管理员拥有基础权限
		basePaths := []string{
			"/base/login",
			"/base/logout",
			"/base/refreshToken",
			"/base/sendcode",
			"/base/changePwd",
			"/base/dashboard",
			"/user/info",
			"/user/resetTotpSecret",
			"/user/changePwd",
			"/menu/access/tree",
			"/sitenav/list",
			"/notice/getall",
		}

		if funk.ContainsString(basePaths, api.Path) {
			roleCasbin = append(roleCasbin, accountModel.RoleCasbin{
				Keyword: roles[1].Keyword,
				Path:    api.Path,
				Method:  api.Method,
			})
		}
	}
	rules := make([][]string, 0)
	for _, c := range roleCasbin {
		rules = append(rules, []string{
			c.Keyword, c.Path, c.Method,
		})
	}
	isAdd, err := global.CasbinEnforcer.AddPolicies(rules)
	if !isAdd {
		global.Log.Errorf("写入casbin数据失败：%v", err)
	}

	// 6.写入分组
	groups := []accountModel.Group{}
	// groups = append(groups, accountModel.Group{
	// 	{
	// 		Model:              gorm.Model{ID: 1},
	// 		GroupName:          "root",
	// 		Remark:             "Base",
	// 		Creator:            "system",
	// 		GroupType:          "",
	// 		ParentId:           0,
	// 		SourceDeptId:       "0",
	// 		Source:             "openldap",
	// 		SourceDeptParentId: "0",
	// 		GroupDN:            config.Conf.Ldap.BaseDN,
	// 	},
	// }

	if config.Conf.DingTalk != nil && config.Conf.DingTalk.Flag != "" {
		groups = append(groups, accountModel.Group{
			GroupName:          config.Conf.DingTalk.Flag,
			Remark:             config.Conf.DingTalk.Flag,
			Creator:            "system",
			GroupType:          "ou",
			ParentId:           0,
			SourceDeptId:       fmt.Sprintf("%s_%d", config.Conf.DingTalk.Flag, 0),
			Source:             config.Conf.DingTalk.Flag,
			SourceDeptParentId: fmt.Sprintf("%s_%d", config.Conf.DingTalk.Flag, 0),
			GroupDN:            fmt.Sprintf("ou=%s,%s", config.Conf.DingTalk.Flag+"root", config.Conf.Ldap.BaseDN),
		})
	}

	if config.Conf.WeCom != nil && config.Conf.WeCom.Flag != "" {
		groups = append(groups, accountModel.Group{
			GroupName:          config.Conf.WeCom.Flag,
			Remark:             config.Conf.WeCom.Flag,
			Creator:            "system",
			GroupType:          "ou",
			ParentId:           0,
			SourceDeptId:       fmt.Sprintf("%s_%d", config.Conf.WeCom.Flag, 0),
			Source:             config.Conf.WeCom.Flag,
			SourceDeptParentId: fmt.Sprintf("%s_%d", config.Conf.WeCom.Flag, 0),
			GroupDN:            fmt.Sprintf("ou=%s,%s", config.Conf.WeCom.Flag+"root", config.Conf.Ldap.BaseDN),
		})
	}

	if config.Conf.FeiShu != nil && config.Conf.FeiShu.Flag != "" {
		groups = append(groups, accountModel.Group{
			GroupName:          config.Conf.FeiShu.Flag,
			Remark:             config.Conf.FeiShu.Flag,
			Creator:            "system",
			GroupType:          "ou",
			ParentId:           0,
			SourceDeptId:       fmt.Sprintf("%s_%d", config.Conf.FeiShu.Flag, 0),
			Source:             config.Conf.FeiShu.Flag,
			SourceDeptParentId: fmt.Sprintf("%s_%d", config.Conf.FeiShu.Flag, 0),
			GroupDN:            fmt.Sprintf("ou=%s,%s", config.Conf.FeiShu.Flag+"root", config.Conf.Ldap.BaseDN),
		})
	}

	// newGroups := make([]accountModel.Group, 0)
	// for _, group := range groups {
	// 	err := global.DB.First(&group, group.ID).Error
	// 	if errors.Is(err, gorm.ErrRecordNotFound) {
	// 		newGroups = append(newGroups, group)
	// 	}
	// }
	// if len(newGroups) > 0 {
	// 	err := global.DB.Create(&newGroups).Error
	// 	if err != nil {
	// 		global.Log.Errorf("写入分组数据失败：%s", err)
	// 	} else {
	// 		global.Log.Infof("写入分组数据成功: %v", newGroups)
	// 	}
	// }
	err = global.DB.Create(&groups).Error
	if err != nil {
		global.Log.Errorf("写入分组数据失败：%s", err)
	} else {
		global.Log.Info("写入分组数据成功")
	}

	sitenavGroup := sitenavModel.NavGroup{
		Model:   gorm.Model{ID: 1},
		Name:    "project01",
		Title:   "项目01",
		Creator: "System",
	}
	if err := global.DB.Create(&sitenavGroup).Error; err != nil {
		global.Log.Errorf("写入导航页分组数据失败：%s", err)
	} else {
		global.Log.Info("写入导航页分组数据成功")
	}

	sitenavSite := sitenavModel.NavSite{
		Name:        "jenkins",
		NavGroupID:  1,
		Description: "project01 Jenkins",
		IconUrl:     "/ui/assets/logo/jenkins.png",
		Link:        "http://127.0.0.1:8080/",
		DocUrl:      "https://www.jenkins.io/doc/",
		Creator:     "System",
	}
	if err := global.DB.Create(&sitenavSite).Error; err != nil {
		global.Log.Errorf("写入导航页站点数据失败：%s", err)
	} else {
		global.Log.Info("写入导航页站点数据成功")
	}

	global.Log.Info("初始化基础数据完成!")
}
