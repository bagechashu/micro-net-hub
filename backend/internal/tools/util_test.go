package tools

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"

	"micro-net-hub/internal/config"
)

func TestGenPass(t *testing.T) {
	pubPEM, privPEM := testRSAKeyPair(t)

	// 注入测试密钥到全局配置, 结束后恢复, 避免影响其它用例
	previousSystem := config.Conf.System
	config.Conf.System = &config.System{
		RSAPublicBytes:  pubPEM,
		RSAPrivateBytes: privPEM,
	}
	t.Cleanup(func() { config.Conf.System = previousSystem })

	const raw = "123456"
	encrypted := NewGenPasswd(raw)
	if encrypted == "" {
		t.Fatal("加密结果为空")
	}
	if encrypted == raw {
		t.Fatalf("加密结果不应等于明文: %q", encrypted)
	}

	if decrypted := NewParsePasswd(encrypted); decrypted != raw {
		t.Fatalf("加解密往返不一致: got %q, want %q", decrypted, raw)
	}
}

// testRSAKeyPair 生成一对用于测试的 RSA 密钥, 返回公钥/私钥的 PEM 编码字节.
//
// 公钥采用 PKIX 编码、私钥采用 PKCS#1 编码, 分别与 RSAEncrypt / RSADecrypt 的
// 解析方式保持一致.
func testRSAKeyPair(t *testing.T) (pubPEM, privPEM []byte) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成测试 RSA 私钥失败: %v", err)
	}

	privPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("序列化测试 RSA 公钥失败: %v", err)
	}
	pubPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	})

	return pubPEM, privPEM
}

func TestArrUintCmp(t *testing.T) {
	a := []uint{1, 2, 3, 4, 6, 9}
	b := []uint{1, 2, 3, 4, 5, 6, 7}
	c, d := ArrUintCmp(a, b)
	fmt.Printf("%v\n", c)
	fmt.Printf("%v\n", d)
}

func TestSliceToString(t *testing.T) {
	a := []uint{1}
	fmt.Printf("%s\n", SliceToString(a, ","))
}

func TestEncodePass(t *testing.T) {
	// to encode a password into ssha
	hashed := EncodePass([]byte("testpass"))
	fmt.Println(string(hashed))
	// to validate a password against saved hash.
	if Matches([]byte(hashed), []byte("testpass")) {
		fmt.Println("Its a match.")
	} else {
		fmt.Println("its not match")
	}
}
