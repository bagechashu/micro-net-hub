package tools

import (
	"crypto/rand"
	"errors"
	"math/big"
	"micro-net-hub/internal/config"
	"unicode"
)

var ErrPasswordNotComplex = errors.New("password must at least 8 characters long, and must contains at least 3 of the following: uppercase letters, lowercase letters, numbers, and symbols")

// 密码加密 使用自适应hash算法, 不可逆
// func GenPasswd(passwd string) string {
// 	hashPasswd, _ := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
// 	return string(hashPasswd)
// }

// 通过比较两个字符串hash判断是否出自同一个明文
// hashPasswd 需要对比的密文
// passwd 明文
// func ComparePasswd(hashPasswd string, passwd string) error {
// 	// if err := bcrypt.CompareHashAndPassword([]byte(hashPasswd), []byte(passwd)); err != nil {
// 	// 	return err
// 	// }

// 	return nil
// }

// 后端密码加密
func NewGenPasswd(passwd string) string {
	// global.Log.Debugf("password of new user: %s", passwd)
	pass, _ := RSAEncrypt([]byte(passwd), config.Conf.System.RSAPublicBytes)
	return string(pass)
}

// 后端密码解密
func NewParsePasswd(passwd string) string {
	pass, _ := RSADecrypt([]byte(passwd), config.Conf.System.RSAPrivateBytes)
	return string(pass)
}

// CheckPasswordComplexity checks if the password meets the complexity requirements:
// - At least 8 characters long
// - Contains at least 3 of the following: uppercase letters, lowercase letters, numbers, and symbols.
func CheckPasswordComplexity(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasNumber, hasSymbol bool

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsDigit(char) {
			hasNumber = true
		} else if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
			hasSymbol = true
		}

		// 一旦发现至少三种类型的字符，立即返回true
		if hasUpper && hasLower && hasNumber {
			return true
		} else if hasUpper && hasLower && hasSymbol {
			return true
		} else if hasUpper && hasNumber && hasSymbol {
			return true
		} else if hasLower && hasNumber && hasSymbol {
			return true
		}
	}

	return false
}

func GeneratePassword(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	password := make([]byte, length)

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "NotRandomPass"
		}
		password[i] = charset[randomIndex.Int64()]
	}

	return string(password)
}
