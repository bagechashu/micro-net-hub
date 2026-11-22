package test

import (
	"strings"
	"testing"

	"micro-net-hub/internal/global"
)

// TestCasbinPoliciesFromDB 只读检查：用当前 casbin 版本加载数据库里的全部策略，
// 每条策略都应能放行；未授权的角色、不存在的接口应被拒绝。需要本地 MySQL（同 TestUserExist）。
func TestCasbinPoliciesFromDB(t *testing.T) {
	InitConfig()

	policies, err := global.CasbinEnforcer.GetPolicy()
	if err != nil {
		t.Fatal(err)
	}
	if len(policies) == 0 {
		t.Skip("no casbin policies in database")
	}
	roles := map[string]bool{}
	for _, p := range policies {
		roles[p[0]] = true
		// 路径参数（如 :id）换成具体值，模拟真实请求
		obj := p[1]
		if strings.Contains(obj, ":") {
			parts := strings.Split(obj, "/")
			for i, s := range parts {
				if strings.HasPrefix(s, ":") {
					parts[i] = "42"
				}
			}
			obj = strings.Join(parts, "/")
		}
		act := p[2]
		if act == "*" {
			act = "GET"
		}
		ok, err := global.CasbinEnforcer.Enforce(p[0], obj, act)
		if err != nil || !ok {
			t.Errorf("policy %v should allow %s %s, got %v %v", p, act, obj, ok, err)
		}
	}
	for _, r := range [][]string{
		{"no-such-role", policies[0][1], policies[0][2]},
		{policies[0][0], "/api/no/such/endpoint", "GET"},
	} {
		if ok, _ := global.CasbinEnforcer.Enforce(r[0], r[1], r[2]); ok {
			t.Errorf("%v should be denied", r)
		}
	}
	t.Logf("checked %d policies for %d roles", len(policies), len(roles))
}
