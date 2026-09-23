package radiusappr

import (
	"strings"

	"micro-net-hub/internal/config"
	accountModel "micro-net-hub/internal/module/account/model"
)

// InApprovalScope 判断用户是否落在审批适用范围内.
//
// 判定规则:
//   - exclude-users 优先级最高, 命中即排除;
//   - users / roles / groups 三个维度之间是"或"关系, 任一维度非空且命中即在范围内
//     (便于用 roles 批量圈定, 再用 users 追加个别人);
//   - 三个维度都为空表示全量生效.
func InApprovalScope(u *accountModel.User, scope config.ApprovalScope) bool {
	if u == nil {
		return false
	}
	if containsFold(scope.ExcludeUsers, u.Username) {
		return false
	}
	if len(scope.Users) == 0 && len(scope.Roles) == 0 && len(scope.Groups) == 0 {
		return true
	}
	if containsFold(scope.Users, u.Username) {
		return true
	}
	for _, role := range u.Roles {
		if role != nil && containsFold(scope.Roles, role.Keyword) {
			return true
		}
	}
	for _, group := range u.Groups {
		if group == nil {
			continue
		}
		if containsFold(scope.Groups, group.GroupName) || containsFold(scope.Groups, group.GroupDN) {
			return true
		}
	}
	return false
}

// containsFold 忽略大小写地判断白名单是否包含目标值
func containsFold(list []string, target string) bool {
	if target == "" {
		return false
	}
	for _, item := range list {
		if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(target)) {
			return true
		}
	}
	return false
}
