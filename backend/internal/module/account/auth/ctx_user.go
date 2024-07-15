package auth

import (
	"errors"

	"micro-net-hub/internal/module/account/model"

	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"
)

// GetCurrentLoginUser 获取当前登录用户信息
// 需要缓存，减少数据库访问
func GetCtxLoginUser(c *gin.Context) (model.User, error) {
	// 通过上下文安全获取用户信息
	ctxUser, exist := c.Get(customClaimsName)
	if !exist {
		return model.User{}, errors.New("用户未登录")
	}

	userInfo, ok := ctxUser.(model.User)
	if !ok {
		return model.User{}, errors.New("用户信息类型错误")
	}

	// 使用明确命名提高可读性，同时减少全局状态的直接依赖
	userCache := model.CacheUserInfoGet(userInfo.Username)
	if userCache != nil {
		// global.Log.Debug("userCache:", userCache)
		return *userCache, nil
	}

	// global.Log.Debug("cache not shoot, query from database")
	// 获取数据库中的用户信息，并处理可能的错误
	var dbUser model.User
	if err := dbUser.GetUserById(userInfo.ID); err != nil {
		return model.User{}, err
	}

	// 更新缓存中的用户信息
	model.CacheUserInfoSet(userInfo.Username, &dbUser)

	return dbUser, nil
}

// GetCurrentUserMinRoleSort  获取当前用户角色排序最小值（最高等级角色）以及当前用户信息
func GetCtxLoginUserMinRole(c *gin.Context) (uint, model.User, error) {
	// 获取当前用户
	ctxUser, err := GetCtxLoginUser(c)
	if err != nil {
		return 0, model.User{}, err
	}

	currentRoles := ctxUser.Roles

	// 初始化currentRoleSorts时预分配切片容量以提高性能
	currentRoleSorts := make([]int, 0, len(currentRoles))
	for _, role := range currentRoles {
		currentRoleSorts = append(currentRoleSorts, int(role.Sort))
	}

	// 处理currentRoles为空的情况，避免外部库对空切片的操作
	if len(currentRoleSorts) == 0 {
		return 0, ctxUser, errors.New("current user donot have roles")
	}

	currentRoleSortMin := uint(funk.MinInt(currentRoleSorts))

	return currentRoleSortMin, ctxUser, nil
}
