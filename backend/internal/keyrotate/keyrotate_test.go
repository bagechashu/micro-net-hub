package keyrotate

import (
	"strings"
	"testing"

	accountModel "micro-net-hub/internal/module/account/model"
	totpModel "micro-net-hub/internal/module/totp/model"
	"micro-net-hub/internal/tools"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&accountModel.User{}, &totpModel.Totp{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func encrypt(t *testing.T, plain string, pub []byte) string {
	t.Helper()
	enc, err := tools.RSAEncrypt([]byte(plain), pub)
	if err != nil {
		t.Fatal(err)
	}
	return string(enc)
}

func decrypt(t *testing.T, cipher string, priv []byte) string {
	t.Helper()
	dec, err := tools.RSADecrypt([]byte(cipher), priv)
	if err != nil {
		t.Fatal(err)
	}
	return string(dec)
}

func TestRotateUsersAndConfig(t *testing.T) {
	oldPriv, oldPub, err := tools.GenerateRSAKeyPair(1024)
	if err != nil {
		t.Fatal(err)
	}
	db := newDB(t)
	users := []accountModel.User{
		{Username: "alice", Mail: "a@x", Password: encrypt(t, "Alice#Pass1", oldPub)},
		{Username: "bob", Mail: "b@x", Password: encrypt(t, "", oldPub)}, // 空口令的密文（LDAP 同步用户）
		{Username: "carol", Mail: "c@x", Password: ""},                   // 未设置口令
		{Username: "dave", Mail: "d@x", Password: "not-a-valid-cipher"},  // 旧私钥解不开
		{Username: "erin", Mail: "e@x", Password: encrypt(t, "Erin#Pass2", oldPub)},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatal(err)
	}
	// 软删除的用户也要处理
	if err := db.Delete(&users[4]).Error; err != nil {
		t.Fatal(err)
	}

	adminPass := encrypt(t, "admin_pass", oldPub)
	cfg := []byte("ldap:\n  # 注释保留\n  admin-pass: \"" + adminPass + "\"\n  user-init-password: \"plain:Init#123\"\n")
	secrets := []configSecret{{"ldap.admin-pass", adminPass}, {"ldap.user-init-password", "plain:Init#123"}}

	plan, err := BuildPlan(db, oldPriv, oldPub, cfg, secrets)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Users) != 3 || plan.EmptyUsers != 1 || len(plan.FailedUsers) != 1 || plan.FailedUsers[0] != "dave" {
		t.Fatalf("unexpected plan: users=%d empty=%d failed=%v", len(plan.Users), plan.EmptyUsers, plan.FailedUsers)
	}
	if err := ApplyUsers(db, plan); err != nil {
		t.Fatal(err)
	}

	var got []accountModel.User
	if err := db.Unscoped().Order("id").Find(&got).Error; err != nil {
		t.Fatal(err)
	}
	if decrypt(t, got[0].Password, plan.NewPrivPEM) != "Alice#Pass1" {
		t.Error("alice password changed")
	}
	if decrypt(t, got[1].Password, plan.NewPrivPEM) != "" {
		t.Error("bob (empty password cipher) should stay empty after rotation")
	}
	if got[2].Password != "" || got[3].Password != "not-a-valid-cipher" {
		t.Error("empty / undecryptable passwords must be left untouched")
	}
	if decrypt(t, got[4].Password, plan.NewPrivPEM) != "Erin#Pass2" {
		t.Error("soft-deleted user should be rotated too")
	}
	if _, err := tools.RSADecrypt([]byte(got[0].Password), oldPriv); err == nil {
		t.Error("old private key should no longer decrypt rotated passwords")
	}

	// config.yml：只替换加密项，注释和 plain: 写法保留
	newCfg := string(plan.ConfigContent)
	if strings.Contains(newCfg, adminPass) || !strings.Contains(newCfg, "# 注释保留") || !strings.Contains(newCfg, "plain:Init#123") {
		t.Fatalf("config not rewritten as expected:\n%s", newCfg)
	}
	start := strings.Index(newCfg, "admin-pass: \"") + len("admin-pass: \"")
	newAdmin := newCfg[start : start+strings.Index(newCfg[start:], "\"")]
	if decrypt(t, newAdmin, plan.NewPrivPEM) != "admin_pass" {
		t.Error("admin-pass not re-encrypted with the new key")
	}
	if len(plan.ConfigSecrets) != 1 {
		t.Errorf("only the encrypted secret should be rotated, got %v", plan.ConfigSecrets)
	}
}

func TestBuildPlanRefusesAmbiguousConfig(t *testing.T) {
	oldPriv, oldPub, _ := tools.GenerateRSAKeyPair(1024)
	db := newDB(t)
	same := encrypt(t, "x", oldPub)
	cfg := []byte("a: " + same + "\nb: " + same + "\n")
	_, err := BuildPlan(db, oldPriv, oldPub, cfg, []configSecret{{"a", same}})
	if err == nil || !strings.Contains(err.Error(), "出现了 2 次") {
		t.Fatalf("expected ambiguity error, got %v", err)
	}
}

func TestBuildPlanRejectsMismatchedKeys(t *testing.T) {
	privA, _, _ := tools.GenerateRSAKeyPair(1024)
	_, pubB, _ := tools.GenerateRSAKeyPair(1024)
	if _, err := BuildPlan(newDB(t), privA, pubB, nil, nil); err == nil {
		t.Fatal("mismatched key pair must be rejected")
	}
}
