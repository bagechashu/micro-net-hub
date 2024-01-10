package model

import (
	"fmt"
	"sync"
	"time"

	"github.com/xlzd/gotp"
	"gorm.io/gorm"
)

type Totp struct {
	gorm.Model
	UserID uint   `gorm:"type:bigint(20);not null;comment:'用户id'" json:"userId"`
	Secret string `gorm:"type:int;not null;unique;comment:'totp secret'" json:"secret"`
}

func (t *Totp) SetTotp() {
	if t.Secret == "" {
		t.Secret = gotp.RandomSecret(32)
	}
}

var (
	userOtpMux = sync.Mutex{}
	userOtp    = map[string]time.Time{}
)

// 判断令牌信息
func CheckOtp(name, otp, secret string) bool {
	key := fmt.Sprintf("%s:%s", name, otp)

	userOtpMux.Lock()
	defer userOtpMux.Unlock()

	// 令牌只能使用一次
	if _, ok := userOtp[key]; ok {
		// 已经存在
		return false
	}
	userOtp[key] = time.Now()

	totp := gotp.NewDefaultTOTP(secret)
	unix := time.Now().Unix()
	verify := totp.Verify(otp, unix)

	return verify
}
