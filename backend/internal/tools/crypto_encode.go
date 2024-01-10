package tools

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

// RSA加密
func RSAEncrypt(data, publicBytes []byte) ([]byte, error) {
	var res []byte
	// 解析公钥
	block, _ := pem.Decode(publicBytes)

	if block == nil {
		return res, fmt.Errorf("无法加密, 公钥可能不正确")
	}

	// 使用X509将解码之后的数据 解析出来
	// x509.MarshalPKCS1PublicKey(block):解析之后无法用，所以采用以下方法：ParsePKIXPublicKey
	keyInit, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return res, fmt.Errorf("无法加密, 公钥可能不正确, %v", err)
	}
	// 使用公钥加密数据
	pubKey := keyInit.(*rsa.PublicKey)
	res, err = rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
	if err != nil {
		return res, fmt.Errorf("无法加密, 公钥可能不正确, %v", err)
	}
	// 将数据加密为base64格式
	return []byte(Base64Encode(res)), nil
}

// 对数据进行解密操作
func RSADecrypt(base64Data, privateBytes []byte) ([]byte, error) {
	var res []byte
	// 将base64数据解析
	data, _ := Base64Decode(string(base64Data))
	// 解析私钥
	block, _ := pem.Decode(privateBytes)
	if block == nil {
		return res, fmt.Errorf("无法解密, 私钥可能不正确,解析私钥失败")
	}
	// 还原数据
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return res, fmt.Errorf("无法解密, 私钥可能不正确,解析PKCS失败 %v", err)
	}
	res, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
	if err != nil {
		return res, fmt.Errorf("无法解密, 私钥可能不正确,解密PKCS1v15失败 %v", err)
	}
	return res, nil
}

// 加密base64字符串
func Base64Encode(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

// 解密base64字符串
func Base64Decode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

func Base32Encode(src []byte) string {
	return base32.StdEncoding.EncodeToString(src)
}

func Base32Decode(str string) ([]byte, error) {
	return base32.StdEncoding.DecodeString(str)
}

func Int64ToBytes(value int64) []byte {
	var result []byte
	mask := int64(0xFF)
	shifts := [8]uint16{56, 48, 40, 32, 24, 16, 8, 0}
	for _, shift := range shifts {
		result = append(result, byte((value>>shift)&mask))
	}
	return result
}
func BytesToUint32(bts []byte) uint32 {
	return (uint32(bts[0]) << 24) + (uint32(bts[1]) << 16) +
		(uint32(bts[2]) << 8) + uint32(bts[3])
}

func HexEncode(random []byte) string {
	return fmt.Sprintf("%x", random)
}

func HexDecode(hexString string) ([]byte, error) {
	return hex.DecodeString(hexString)
}

func HmacSha1Hash(key, data []byte) []byte {
	h := hmac.New(sha1.New, key)
	if total := len(data); total > 0 {
		h.Write(data)
	}
	return h.Sum(nil)
}
func GenerateRandom(length int) []byte {
	str := make([]byte, length)
	_, err := rand.Read(str)
	if err != nil {
		fmt.Println("Error generating random bytes:", err)
		return str
	}
	return str
}
