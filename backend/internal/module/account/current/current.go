package current

import (
	"errors"

	accountModel "micro-net-hub/internal/module/account/model"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/thoas/go-funk"
)

// GetCurrentUserMinRoleSort  获取当前用户角色排序最小值（最高等级角色）以及当前用户信息
func GetCurrentUserMinRoleSort(c *gin.Context) (uint, accountModel.User, error) {
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
	currentRoleSortMin := uint(funk.MinInt(currentRoleSorts))

	return currentRoleSortMin, ctxUser, nil
}

// GetCurrentLoginUser 获取当前登录用户信息
// 需要缓存，减少数据库访问
func GetCurrentLoginUser(c *gin.Context) (accountModel.User, error) {
	var newUser accountModel.User
	ctxUser, exist := c.Get("user")
	if !exist {
		return newUser, errors.New("用户未登录")
	}
	u, _ := ctxUser.(accountModel.User)

	// 先获取缓存
	cacheUser, found := accountModel.UserInfoCache.Get(u.Username)
	var user accountModel.User
	var err error
	if found {
		user = cacheUser.(accountModel.User)
		err = nil
	} else {
		// 缓存中没有就获取数据库
		err = user.GetUserById(u.ID)
		// 获取成功就缓存
		if err != nil {
			accountModel.UserInfoCache.Delete(u.Username)
		} else {
			accountModel.UserInfoCache.Set(u.Username, user, cache.DefaultExpiration)
		}
	}
	return user, err
}
