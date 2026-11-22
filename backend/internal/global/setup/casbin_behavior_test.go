package setup

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"micro-net-hub/internal/config"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestCasbinBehavior 用内存 SQLite 跑一遍项目里用到的 casbin 接口（与线上相同的 RBAC 模型和 gorm-adapter），
// 记录每一步的返回值和策略结果。升级 casbin 时可以把新旧版本的记录对比，确认行为一致：
//
//	CASBIN_TRANSCRIPT=/tmp/casbin.txt go test ./internal/global/setup -run TestCasbinBehavior
func TestCasbinBehavior(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	newEnforcer := func() *casbin.Enforcer {
		a, err := gormadapter.NewAdapterByDB(db)
		if err != nil {
			t.Fatal(err)
		}
		m, err := model.NewModelFromString(config.RBAC_MODEL)
		if err != nil {
			t.Fatal(err)
		}
		e, err := casbin.NewEnforcer(m, a)
		if err != nil {
			t.Fatal(err)
		}
		if err := e.LoadPolicy(); err != nil {
			t.Fatal(err)
		}
		return e
	}

	var out strings.Builder
	logf := func(format string, args ...any) { fmt.Fprintf(&out, format+"\n", args...) }
	policies := func(e *casbin.Enforcer) string {
		p, err := e.GetPolicy()
		if err != nil {
			return "err: " + err.Error()
		}
		rows := make([]string, 0, len(p))
		for _, r := range p {
			rows = append(rows, strings.Join(r, ","))
		}
		sort.Strings(rows)
		return strings.Join(rows, " | ")
	}

	e := newEnforcer()
	a := []string{"admin", "/api/user/:id", "GET"}
	b := []string{"admin", "/api/user/list", "GET"}
	c := []string{"admin", "/api/base/*", "*"}
	d := []string{"user", "/api/menu/list", "GET"}

	ok, err := e.AddPolicies([][]string{a, b})
	logf("AddPolicies(a,b) = %v %v -> %s", ok, err, policies(e))
	ok, err = e.AddPolicies([][]string{b, c})
	logf("AddPolicies(b,c) [b exists] = %v %v -> %s", ok, err, policies(e))
	ok, err = e.AddPolicies([][]string{c, d})
	logf("AddPolicies(c,d) = %v %v -> %s", ok, err, policies(e))
	ok, err = e.RemovePolicies([][]string{d, {"nobody", "/x", "GET"}})
	logf("RemovePolicies(d,missing) [one missing] = %v %v -> %s", ok, err, policies(e))
	ok, err = e.RemovePolicies([][]string{d})
	logf("RemovePolicies(d) = %v %v -> %s", ok, err, policies(e))
	ok, err = e.UpdatePolicies([][]string{a}, [][]string{{"admin", "/api/user/:id", "PUT"}})
	logf("UpdatePolicies(a -> PUT) = %v %v -> %s", ok, err, policies(e))

	f, err := e.GetFilteredPolicy(0, "admin")
	logf("GetFilteredPolicy(0,admin) = %d rows %v", len(f), err)
	f, err = e.GetFilteredPolicy(1, "/api/user/list")
	logf("GetFilteredPolicy(1,/api/user/list) = %v %v", f, err)
	f, err = e.GetFilteredPolicy(0, "nobody")
	logf("GetFilteredPolicy(0,nobody) = %v (nil=%v) %v", f, f == nil, err)

	for _, r := range [][]string{
		{"admin", "/api/user/5", "PUT"},
		{"admin", "/api/user/5", "GET"},
		{"admin", "/api/user/list", "GET"},
		{"admin", "/api/base/login", "POST"},
		{"admin", "/api/base/a/b", "DELETE"},
		{"user", "/api/user/list", "GET"},
		{"admin", "/api/menu/list", "GET"},
	} {
		allowed, err := e.Enforce(r[0], r[1], r[2])
		logf("Enforce(%s) = %v %v", strings.Join(r, ","), allowed, err)
	}

	// 自动保存：新的 enforcer 从同一个库加载，应看到相同的策略
	logf("reload -> %s", policies(newEnforcer()))

	if p := os.Getenv("CASBIN_TRANSCRIPT"); p != "" {
		if err := os.WriteFile(p, []byte(out.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("\n" + out.String())
}
