package tools

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"slices"
	"strings"
)

// RSAKeyBits 新生成密钥的长度
const RSAKeyBits = 2048

// PlainSecretPrefix 配置文件里的口令（ldap.admin-pass、ldap.user-init-password）以此前缀开头时按明文使用，
// 否则按 RSA 密文解密。部署时自动生成的新密钥无法解开旧样例里的密文，所以样例改用明文写法。
const PlainSecretPrefix = "plain:"

// BundledPublicKeyFingerprints 仓库里提交过的公开 RSA 公钥（SHA-256 指纹）。
// 生产环境仍在使用它们时，任何人都能用仓库里的私钥解出所有用户密码，需要执行 rotate-rsa-key。
var BundledPublicKeyFingerprints = []string{
	"371ca52bac8d1c49e99bada4d4754d4f92ce57e7fb0943e6536c8d40cc8c37e3",
}

// GenerateRSAKeyPair 生成 RSA 密钥对。格式与 RSAEncrypt / RSADecrypt 的解析方式一致：
// 私钥为 PKCS#1（BEGIN RSA PRIVATE KEY），公钥为 PKIX（BEGIN PUBLIC KEY）。
func GenerateRSAKeyPair(bits int) (privPEM, pubPEM []byte, err error) {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("生成 RSA 密钥失败: %w", err)
	}
	privPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("编码 RSA 公钥失败: %w", err)
	}
	pubPEM = pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	return privPEM, pubPEM, nil
}

// PublicKeyFingerprint 公钥（PKIX DER）的 SHA-256 十六进制指纹
func PublicKeyFingerprint(pubPEM []byte) (string, error) {
	block, _ := pem.Decode(pubPEM)
	if block == nil {
		return "", fmt.Errorf("无法解析公钥 PEM")
	}
	if _, err := x509.ParsePKIXPublicKey(block.Bytes); err != nil {
		return "", fmt.Errorf("无法解析公钥: %w", err)
	}
	sum := sha256.Sum256(block.Bytes)
	return hex.EncodeToString(sum[:]), nil
}

// IsBundledPublicKey 是否为仓库里提交过的公开密钥
func IsBundledPublicKey(pubPEM []byte) bool {
	fp, err := PublicKeyFingerprint(pubPEM)
	if err != nil {
		return false
	}
	return slices.Contains(BundledPublicKeyFingerprints, fp)
}

// CheckRSAKeyPair 用一段测试数据加解密一次，确认公钥和私钥属于同一对
func CheckRSAKeyPair(privPEM, pubPEM []byte) error {
	probe := []byte("micro-net-hub-rsa-key-check")
	enc, err := RSAEncrypt(probe, pubPEM)
	if err != nil {
		return err
	}
	dec, err := RSADecrypt(enc, privPEM)
	if err != nil {
		return err
	}
	if !bytes.Equal(dec, probe) {
		return fmt.Errorf("公钥和私钥不匹配")
	}
	return nil
}

// ReencryptPassword 用旧私钥解出明文、再用新公钥加密，并用新私钥回解校验
func ReencryptPassword(cipher string, oldPriv, newPub, newPriv []byte) (string, error) {
	plain, err := RSADecrypt([]byte(cipher), oldPriv)
	if err != nil {
		return "", err
	}
	enc, err := RSAEncrypt(plain, newPub)
	if err != nil {
		return "", err
	}
	check, err := RSADecrypt(enc, newPriv)
	if err != nil || !bytes.Equal(check, plain) {
		return "", fmt.Errorf("新密钥回解校验失败: %v", err)
	}
	return string(enc), nil
}

// ParseConfigSecret 解析配置文件里的口令：plain: 前缀为明文，否则按 RSA 密文用当前私钥解密
func ParseConfigSecret(v string) string {
	if plain, ok := strings.CutPrefix(v, PlainSecretPrefix); ok {
		return plain
	}
	return NewParsePasswd(v)
}

// IsEncryptedConfigSecret 配置口令是否为 RSA 密文（非空且不是 plain: 写法）
func IsEncryptedConfigSecret(v string) bool {
	return v != "" && !strings.HasPrefix(v, PlainSecretPrefix)
}
