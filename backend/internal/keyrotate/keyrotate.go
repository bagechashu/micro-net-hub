// Package keyrotate 轮换 RSA 密钥：生成新密钥对，把数据库里用户口令的密文和 config.yml 里加密的口令
// 用旧私钥解出后用新公钥重新加密。用户口令本身不变，用户无需重新设置；LDAP 里的口令（SSHA）也不受影响。
package keyrotate

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"micro-net-hub/internal/config"
	accountModel "micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/tools"

	"gorm.io/gorm"
)

// Options 轮换参数
type Options struct {
	DryRun     bool      // 只统计、不修改
	ConfigPath string    // config.yml 路径（重新加密其中的口令）
	Out        io.Writer // 输出
}

// userCipher 需要重新加密的用户口令
type userCipher struct {
	ID       uint
	Username string
	Password string
}

// Plan 轮换计划：全部在内存里算好、校验通过后才落盘
type Plan struct {
	NewPrivPEM, NewPubPEM []byte
	Users                 map[uint]string // 用户 ID -> 新密文
	FailedUsers           []string        // 旧私钥解不开的用户（原样保留，本来就无法登录）
	EmptyUsers            int             // 口令为空、无需处理的用户
	ConfigContent         []byte          // 重新加密口令后的 config.yml 内容（无需修改时为 nil）
	ConfigSecrets         []string        // 被重新加密的配置项
}

// configSecret 需要重新加密的配置项
type configSecret struct {
	Name  string
	Value string
}

// BuildPlan 生成新密钥并计算所有新密文（不修改任何数据）
func BuildPlan(db *gorm.DB, oldPriv, oldPub []byte, configContent []byte, secrets []configSecret) (*Plan, error) {
	if err := tools.CheckRSAKeyPair(oldPriv, oldPub); err != nil {
		return nil, fmt.Errorf("当前密钥不可用: %w", err)
	}
	newPriv, newPub, err := tools.GenerateRSAKeyPair(tools.RSAKeyBits)
	if err != nil {
		return nil, err
	}
	plan := &Plan{NewPrivPEM: newPriv, NewPubPEM: newPub, Users: map[uint]string{}}

	var users []userCipher
	if err := db.Unscoped().Model(&accountModel.User{}).Select("id", "username", "password").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("读取用户失败: %w", err)
	}
	for _, u := range users {
		if u.Password == "" {
			plan.EmptyUsers++
			continue
		}
		enc, err := tools.ReencryptPassword(u.Password, oldPriv, newPub, newPriv)
		if err != nil {
			plan.FailedUsers = append(plan.FailedUsers, u.Username)
			continue
		}
		plan.Users[u.ID] = enc
	}

	// config.yml 里的加密口令：按原文精确替换，保留文件里的注释和格式
	content := configContent
	for _, s := range secrets {
		if !tools.IsEncryptedConfigSecret(s.Value) {
			continue
		}
		enc, err := tools.ReencryptPassword(s.Value, oldPriv, newPub, newPriv)
		if err != nil {
			return nil, fmt.Errorf("配置项 %s 无法用当前私钥解密: %w", s.Name, err)
		}
		if n := bytes.Count(content, []byte(s.Value)); n != 1 {
			return nil, fmt.Errorf("配置项 %s 的密文在 config.yml 里出现了 %d 次（应为 1 次），无法自动替换", s.Name, n)
		}
		content = bytes.Replace(content, []byte(s.Value), []byte(enc), 1)
		plan.ConfigSecrets = append(plan.ConfigSecrets, s.Name)
	}
	if len(plan.ConfigSecrets) > 0 {
		plan.ConfigContent = content
	}
	return plan, nil
}

// ApplyUsers 在一个事务里写入新的用户口令密文
func ApplyUsers(db *gorm.DB, plan *Plan) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for id, enc := range plan.Users {
			// UpdateColumn：不触发钩子、不改 updated_at
			res := tx.Unscoped().Model(&accountModel.User{}).Where("id = ?", id).UpdateColumn("password", enc)
			if res.Error != nil {
				return fmt.Errorf("更新用户 %d 失败: %w", id, res.Error)
			}
		}
		return nil
	})
}

// Run 执行轮换（需已初始化 config 和数据库）
func Run(db *gorm.DB, opt Options) error {
	out := opt.Out
	if out == nil {
		out = os.Stdout
	}
	sys := config.Conf.System
	privPath, pubPath := sys.RSAPrivateKey, sys.RSAPublicKey
	if len(sys.RSAPrivateBytes) == 0 || len(sys.RSAPublicBytes) == 0 {
		return fmt.Errorf("读取当前密钥失败（%s、%s）", pubPath, privPath)
	}

	var configContent []byte
	if opt.ConfigPath != "" {
		b, err := os.ReadFile(opt.ConfigPath)
		if err != nil {
			return fmt.Errorf("读取 %s 失败: %w", opt.ConfigPath, err)
		}
		configContent = b
	}
	secrets := []configSecret{
		{"ldap.admin-pass", config.Conf.Ldap.AdminPass},
		{"ldap.user-init-password", config.Conf.Ldap.UserInitPassword},
	}

	plan, err := BuildPlan(db, sys.RSAPrivateBytes, sys.RSAPublicBytes, configContent, secrets)
	if err != nil {
		return err
	}

	fp, _ := tools.PublicKeyFingerprint(sys.RSAPublicBytes)
	newFP, _ := tools.PublicKeyFingerprint(plan.NewPubPEM)
	fmt.Fprintf(out, "当前公钥指纹: %s%s\n", fp, map[bool]string{true: "（仓库公开密钥）", false: ""}[tools.IsBundledPublicKey(sys.RSAPublicBytes)])
	fmt.Fprintf(out, "新公钥指纹:   %s（%d 位）\n", newFP, tools.RSAKeyBits)
	fmt.Fprintf(out, "用户口令: 重新加密 %d 个，口令为空跳过 %d 个，无法解密 %d 个\n", len(plan.Users), plan.EmptyUsers, len(plan.FailedUsers))
	if len(plan.FailedUsers) > 0 {
		fmt.Fprintf(out, "  无法解密的用户（原样保留，本来就无法用口令登录，需管理员重设口令）: %s\n", strings.Join(plan.FailedUsers, ", "))
	}
	if len(plan.ConfigSecrets) > 0 {
		fmt.Fprintf(out, "config.yml: 重新加密 %s\n", strings.Join(plan.ConfigSecrets, "、"))
	} else {
		fmt.Fprintln(out, "config.yml: 没有需要重新加密的口令（未配置或使用 plain: 明文写法）")
	}
	// 新密钥的密文更长（2048 位为 344 字符），旧库的 users.password 可能是 varchar(255)，需要先放宽
	widen, colLen, need := passwordColumnTooShort(db, plan)
	if widen {
		fmt.Fprintf(out, "users.password 列长 %d，新密文需要 %d：执行时会先按当前模型放宽该列\n", colLen, need)
	}
	if opt.DryRun {
		fmt.Fprintln(out, "dry-run：没有做任何修改")
		return nil
	}

	// 1. 备份旧密钥、旧配置，并保存一份新密钥（之后任何一步失败都能从这里恢复）
	backupDir := filepath.Join(filepath.Dir(privPath), "rsa-backup-"+time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return fmt.Errorf("创建备份目录失败: %w", err)
	}
	backups := []struct {
		name string
		data []byte
		perm os.FileMode
	}{
		{"old-" + filepath.Base(privPath), sys.RSAPrivateBytes, 0o600},
		{"old-" + filepath.Base(pubPath), sys.RSAPublicBytes, 0o644},
		{"new-" + filepath.Base(privPath), plan.NewPrivPEM, 0o600},
		{"new-" + filepath.Base(pubPath), plan.NewPubPEM, 0o644},
	}
	if configContent != nil {
		backups = append(backups, struct {
			name string
			data []byte
			perm os.FileMode
		}{"old-" + filepath.Base(opt.ConfigPath), configContent, 0o600})
	}
	for _, b := range backups {
		if err := os.WriteFile(filepath.Join(backupDir, b.name), b.data, b.perm); err != nil {
			return fmt.Errorf("写入备份失败: %w", err)
		}
	}
	fmt.Fprintf(out, "已备份旧密钥、旧配置和新密钥到 %s\n", backupDir)

	// 2. 数据库（放宽列是 DDL，在事务之外执行；之后事务失败时列仍保持放宽，不影响使用）
	if widen {
		if err := db.Migrator().AlterColumn(&accountModel.User{}, "Password"); err != nil {
			return fmt.Errorf("放宽 users.password 列失败: %w（数据库未修改，密钥和配置也未修改）", err)
		}
		fmt.Fprintln(out, "已放宽 users.password 列")
	}
	if err := ApplyUsers(db, plan); err != nil {
		return fmt.Errorf("%w（数据库未修改，密钥和配置也未修改）", err)
	}
	fmt.Fprintln(out, "数据库已更新")

	// 3. 密钥文件和 config.yml
	failed := false
	for _, f := range []struct {
		path string
		data []byte
		perm os.FileMode
	}{
		{privPath, plan.NewPrivPEM, 0o600},
		{pubPath, plan.NewPubPEM, 0o644},
	} {
		if err := replaceFile(f.path, f.data, f.perm); err != nil {
			fmt.Fprintf(out, "写入 %s 失败: %v\n", f.path, err)
			failed = true
		}
	}
	if plan.ConfigContent != nil {
		if err := replaceFile(opt.ConfigPath, plan.ConfigContent, 0o600); err != nil {
			fmt.Fprintf(out, "写入 %s 失败: %v\n", opt.ConfigPath, err)
			failed = true
		}
	}
	if failed {
		return fmt.Errorf("数据库已经换成新密钥加密，但部分文件没写成功。请手工把 %s 里的 new-* 文件复制到对应位置（config.yml 里的口令需用新公钥重新加密，或改成 plain: 明文写法）后再启动服务", backupDir)
	}
	fmt.Fprintln(out, "密钥文件和 config.yml 已更新。请启动（或重启）服务使新密钥生效；用户口令不变，无需重新设置")
	return nil
}

// passwordColumnTooShort 判断 users.password 列是否放不下新密文。
// SQLite 不检查字符串长度，不需要也不应该改表（SQLite 改列会重建整张表）。
func passwordColumnTooShort(db *gorm.DB, plan *Plan) (widen bool, colLen int64, need int) {
	for _, enc := range plan.Users {
		need = max(need, len(enc))
	}
	if need == 0 || db.Dialector.Name() == "sqlite" {
		return false, 0, need
	}
	cols, err := db.Migrator().ColumnTypes(&accountModel.User{})
	if err != nil {
		return false, 0, need
	}
	for _, c := range cols {
		if strings.EqualFold(c.Name(), "password") {
			if l, ok := c.Length(); ok && l > 0 && l < int64(need) {
				return true, l, need
			}
		}
	}
	return false, 0, need
}

// replaceFile 替换文件内容：优先写临时文件再改名（原子替换）；
// 目标是 Docker 单文件挂载点时改名会失败，此时直接覆盖写入原文件。
func replaceFile(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp-rotate"
	if err := os.WriteFile(tmp, data, perm); err == nil {
		if err := os.Rename(tmp, path); err == nil {
			return nil
		}
		_ = os.Remove(tmp)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
