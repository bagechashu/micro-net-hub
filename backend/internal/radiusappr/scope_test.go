package radiusappr

import (
	"testing"

	"micro-net-hub/internal/config"
	accountModel "micro-net-hub/internal/module/account/model"

	"github.com/stretchr/testify/assert"
)

// scopeUser 构造带角色与分组的测试用户
func scopeUser() *accountModel.User {
	return &accountModel.User{
		Username: "alice",
		Roles: []*accountModel.Role{
			{Keyword: "user"},
			{Keyword: "dev"},
		},
		Groups: []*accountModel.Group{
			{GroupName: "研发部", GroupDN: "ou=dev,dc=example,dc=com"},
		},
	}
}

// TestInApprovalScope 覆盖全量生效/白名单/黑名单/角色与分组命中
func TestInApprovalScope(t *testing.T) {
	tests := []struct {
		name  string
		scope config.ApprovalScope
		want  bool
	}{
		{name: "各维度为空表示全量生效", scope: config.ApprovalScope{}, want: true},
		{name: "users 命中", scope: config.ApprovalScope{Users: []string{"alice"}}, want: true},
		{name: "users 不命中", scope: config.ApprovalScope{Users: []string{"bob"}}, want: false},
		{name: "roles 命中", scope: config.ApprovalScope{Roles: []string{"dev"}}, want: true},
		{name: "roles 不命中", scope: config.ApprovalScope{Roles: []string{"admin"}}, want: false},
		{name: "groups 按名称命中", scope: config.ApprovalScope{Groups: []string{"研发部"}}, want: true},
		{name: "groups 按 DN 命中", scope: config.ApprovalScope{Groups: []string{"ou=dev,dc=example,dc=com"}}, want: true},
		{name: "groups 不命中", scope: config.ApprovalScope{Groups: []string{"市场部"}}, want: false},
		{name: "多维度之间为或关系", scope: config.ApprovalScope{Roles: []string{"admin"}, Users: []string{"alice"}}, want: true},
		{name: "exclude-users 优先于角色命中", scope: config.ApprovalScope{Roles: []string{"dev"}, ExcludeUsers: []string{"alice"}}, want: false},
		{name: "匹配忽略大小写与空白", scope: config.ApprovalScope{Users: []string{" ALICE "}}, want: true},
	}

	user := scopeUser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, InApprovalScope(user, tt.scope))
		})
	}

	assert.False(t, InApprovalScope(nil, config.ApprovalScope{}), "用户信息为空时不应判定为在范围内")
}
