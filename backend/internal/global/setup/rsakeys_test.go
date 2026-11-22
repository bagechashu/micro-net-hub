package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	accountModel "micro-net-hub/internal/module/account/model"
	totpModel "micro-net-hub/internal/module/totp/model"
	"micro-net-hub/internal/tools"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRSATest(t *testing.T, withUser bool) string {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&accountModel.User{}, &totpModel.Totp{}); err != nil {
		t.Fatal(err)
	}
	if withUser {
		if err := db.Create(&accountModel.User{Username: "u1", Mail: "u1@x"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	oldDB, oldLog, oldSys := global.DB, global.Log, config.Conf.System
	t.Cleanup(func() { global.DB, global.Log, config.Conf.System = oldDB, oldLog, oldSys })
	global.DB, global.Log = db, zap.NewNop().Sugar()

	dir := t.TempDir()
	config.Conf.System = &config.System{
		Mode:          "release",
		RSAPublicKey:  filepath.Join(dir, "keys", "auth-pub.pem"),
		RSAPrivateKey: filepath.Join(dir, "keys", "auth-priv-rsa.pem"),
	}
	return dir
}

func TestInitRSAKeysGeneratesOnFreshInstall(t *testing.T) {
	setupRSATest(t, false)
	if err := initRSAKeys(); err != nil {
		t.Fatal(err)
	}
	sys := config.Conf.System
	if err := tools.CheckRSAKeyPair(sys.RSAPrivateBytes, sys.RSAPublicBytes); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(sys.RSAPrivateKey)
	if err != nil || fi.Mode().Perm() != 0o600 {
		t.Fatalf("private key should be created with 0600, got %v %v", fi.Mode().Perm(), err)
	}
	if tools.IsBundledPublicKey(sys.RSAPublicBytes) {
		t.Fatal("generated key must not be the bundled one")
	}
	// 再次启动：使用已有密钥，不重新生成
	before := string(sys.RSAPublicBytes)
	if err := initRSAKeys(); err != nil || string(config.Conf.System.RSAPublicBytes) != before {
		t.Fatalf("existing keys should be reused: %v", err)
	}
}

func TestInitRSAKeysRefusesWhenUsersExist(t *testing.T) {
	setupRSATest(t, true)
	err := initRSAKeys()
	if err == nil || !strings.Contains(err.Error(), "已有 1 个用户") {
		t.Fatalf("expected refusal, got %v", err)
	}
	if _, statErr := os.Stat(config.Conf.System.RSAPrivateKey); statErr == nil {
		t.Fatal("must not generate keys when users exist")
	}
}

func TestInitRSAKeysRefusesHalfPair(t *testing.T) {
	setupRSATest(t, false)
	_, pub, _ := tools.GenerateRSAKeyPair(1024)
	if err := writeNewFile(config.Conf.System.RSAPublicKey, pub, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := initRSAKeys(); err == nil || !strings.Contains(err.Error(), "只找到一个") {
		t.Fatalf("expected half-pair error, got %v", err)
	}
}

func TestInitRSAKeysRejectsMismatchedPair(t *testing.T) {
	setupRSATest(t, false)
	privA, _, _ := tools.GenerateRSAKeyPair(1024)
	_, pubB, _ := tools.GenerateRSAKeyPair(1024)
	_ = writeNewFile(config.Conf.System.RSAPrivateKey, privA, 0o600)
	_ = writeNewFile(config.Conf.System.RSAPublicKey, pubB, 0o644)
	if err := initRSAKeys(); err == nil || !strings.Contains(err.Error(), "不可用") {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestParseConfigSecret(t *testing.T) {
	if tools.ParseConfigSecret("plain:admin_pass") != "admin_pass" {
		t.Fatal("plain: prefix should be returned as plaintext")
	}
	if tools.IsEncryptedConfigSecret("plain:x") || tools.IsEncryptedConfigSecret("") || !tools.IsEncryptedConfigSecret("abc") {
		t.Fatal("IsEncryptedConfigSecret misclassified")
	}
}
