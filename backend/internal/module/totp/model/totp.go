package model

import (
	"fmt"
	"micro-net-hub/internal/global"
	"micro-net-hub/internal/tools"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Totp struct {
	gorm.Model
	UserID uint   `gorm:"type:bigint(20);not null;comment:'用户id'" json:"userId"`
	Secret string `gorm:"type:varchar(32);unique;comment:'totp secret'" json:"secret"`
}

func (t *Totp) SetTotpSecret() {
	if t.Secret == "" {
		// use base32 (google type) secret
		_, t.Secret = GenerateQRsecret()
	}
}

func (t *Totp) ReSetTotpSecret() {
	_, t.Secret = GenerateQRsecret()
}

func CheckTotp(secret string, totp string) (valid bool) {
	valid = false
	secretUpper := strings.ToUpper(secret)
	secretKey, err := tools.Base32Decode(secretUpper)
	if err != nil {
		return
	}
	code := GetGoogleTotp(secretKey)
	if totp == code {
		valid = true
	}
	global.Log.Debugf("secret: %v, google-code: %v, user-input-totp: %v", secret, code, totp)
	return
}

func GetGoogleTotp(key []byte) string {
	hash := tools.HmacSha1Hash(key, tools.Int64ToBytes(time.Now().Unix()/30))
	offset := hash[len(hash)-1] & 0x0F
	hashParts := hash[offset : offset+4]
	hashParts[0] = hashParts[0] & 0x7F
	number := tools.BytesToUint32(hashParts)
	return fmt.Sprintf("%06d", number%1000000)
}

func GenerateQRsecret() (hexString string, base32String string) {
	random := tools.GenerateRandom(16)
	// fmt.Println(random)

	// // 以下等同于 "echo -n 12345qwertasdfga | xxd -c 256 -ps"
	// random := "12345qwertasdfga"
	// hexString := hexEncode([]byte(random))
	// fmt.Println(hexString)

	// Token key (RFC 4226)
	hexString = tools.HexEncode(random)
	// fmt.Printf("Token key (RFC 4226): %s\n", hexString)

	// Base32 key (Google Authenticator)
	base32String = tools.Base32Encode([]byte(random))
	// fmt.Printf("Base32 key (Google Authenticator): %s\n", base32String)
	return hexString, base32String
}
