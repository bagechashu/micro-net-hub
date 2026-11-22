package setup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	accountModel "micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/tools"
)

// InitRSAKeys 检查并加载 RSA 密钥（用于登录时加密传输口令、可逆存储用户口令）。需在 InitDB 之后调用。
//   - 两个密钥文件都不存在、且数据库里还没有用户（全新部署）：自动生成 2048 位密钥对并写入配置的路径。
//   - 都不存在但已有用户：拒绝启动。多半是漏挂载了密钥文件，此时生成新密钥会让所有已有用户无法登录。
//   - 只存在一个：拒绝启动。
//   - 仍在使用仓库里提交过的公开密钥：release 模式下告警，提示执行 rotate-rsa-key。
func InitRSAKeys() {
	if err := initRSAKeys(); err != nil {
		global.Log.Panicf("初始化 RSA 密钥失败：%v", err)
	}
}

func initRSAKeys() error {
	sys := config.Conf.System
	if sys.RSAPublicKey == "" {
		sys.RSAPublicKey = config.ResolvePath(config.DefaultRSAPublicKeyFile)
	}
	if sys.RSAPrivateKey == "" {
		sys.RSAPrivateKey = config.ResolvePath(config.DefaultRSAPrivateKeyFile)
	}
	pubExists, err := regularFileExists(sys.RSAPublicKey)
	if err != nil {
		return err
	}
	privExists, err := regularFileExists(sys.RSAPrivateKey)
	if err != nil {
		return err
	}

	switch {
	case pubExists && privExists:
		// 正常情况
	case !pubExists && !privExists:
		var users int64
		if err := global.DB.Unscoped().Model(&accountModel.User{}).Count(&users).Error; err != nil {
			return fmt.Errorf("查询用户数量失败: %w", err)
		}
		if users > 0 {
			return fmt.Errorf("找不到 RSA 密钥文件（%s、%s），但数据库里已有 %d 个用户。"+
				"用户口令是用原来的密钥加密保存的，生成新密钥会导致所有人无法登录。"+
				"请检查密钥文件是否挂载或路径配置（system.rsa-public-key / rsa-private-key），恢复原来的密钥后再启动",
				sys.RSAPublicKey, sys.RSAPrivateKey, users)
		}
		priv, pub, err := tools.GenerateRSAKeyPair(tools.RSAKeyBits)
		if err != nil {
			return err
		}
		if err := writeNewFile(sys.RSAPrivateKey, priv, 0o600); err != nil {
			return fmt.Errorf("写入私钥失败: %w", err)
		}
		if err := writeNewFile(sys.RSAPublicKey, pub, 0o644); err != nil {
			return fmt.Errorf("写入公钥失败: %w", err)
		}
		global.Log.Warnf("已生成新的 %d 位 RSA 密钥对：%s、%s。请妥善备份，丢失后已有用户的口令将无法校验",
			tools.RSAKeyBits, sys.RSAPublicKey, sys.RSAPrivateKey)
	default:
		return fmt.Errorf("RSA 密钥文件只找到一个（公钥 %s 存在=%v，私钥 %s 存在=%v），请补齐或恢复原来的密钥对",
			sys.RSAPublicKey, pubExists, sys.RSAPrivateKey, privExists)
	}

	sys.RSAPublicBytes = config.RSAReadKeyFromFile(sys.RSAPublicKey)
	sys.RSAPrivateBytes = config.RSAReadKeyFromFile(sys.RSAPrivateKey)
	if err := tools.CheckRSAKeyPair(sys.RSAPrivateBytes, sys.RSAPublicBytes); err != nil {
		return fmt.Errorf("RSA 密钥不可用（%s、%s）: %w", sys.RSAPublicKey, sys.RSAPrivateKey, err)
	}

	if tools.IsBundledPublicKey(sys.RSAPublicBytes) {
		msg := "当前使用的是仓库里提交过的公开 RSA 密钥，任何人都能用仓库里的私钥解出用户口令。" +
			"请停止服务后执行 `micro-net-hub rotate-rsa-key` 轮换密钥"
		if sys.Mode == "release" {
			global.Log.Warn(msg)
		} else {
			global.Log.Info(msg + "（开发环境可忽略）")
		}
	}
	return nil
}

func regularFileExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("检查 %s 失败: %w", path, err)
	}
	if !fi.Mode().IsRegular() {
		// Docker 挂载不存在的宿主文件时会创建同名目录，这里给出明确提示
		return false, fmt.Errorf("%s 不是普通文件（可能是挂载时被创建成了目录）", path)
	}
	return true, nil
}

// writeNewFile 创建新文件（已存在则报错），必要时创建上级目录
func writeNewFile(path string, data []byte, perm os.FileMode) error {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
