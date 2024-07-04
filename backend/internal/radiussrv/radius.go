/*

Copyright 2020 Andrey Devyatkin.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package radiussrv

import (
	"fmt"
	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	accountModel "micro-net-hub/internal/module/account/model"
	totpModel "micro-net-hub/internal/module/totp/model"
	"time"

	"github.com/patrickmn/go-cache"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type loginAttemptInfo struct {
	Times        int
	LastFailedAt time.Time
}

// 初始化一个缓存实例，设置过期时间为 banDurationMinute 分钟
var banDurationMinute float64 = 5
var loginAttemptsCache = cache.New(time.Duration(banDurationMinute)*time.Minute, time.Duration(banDurationMinute)*time.Minute)

// AuthRequest - encapsulates approval logic
func AuthRequest(username string, password string) (valid bool, err error) {
	valid = false

	var loginAttempt = &loginAttemptInfo{}
	attempt, found := loginAttemptsCache.Get(username)
	if found {
		loginAttempt = attempt.(*loginAttemptInfo)
		global.Log.Debugf("radius cache: %s-before: %+v attempt %+v", username, loginAttempt.Times, loginAttempt.LastFailedAt)

		// 如果最后一次失败尝试距离现在不足10分钟，则返回错误并禁止登录
		if loginAttempt.Times > config.Conf.Radius.FailTimesBeforeBlock5min && time.Since(loginAttempt.LastFailedAt).Minutes() < banDurationMinute {
			err = fmt.Errorf("登录失败次数过多，账户已被锁定5分钟")
			return
		}
	}

	// password 后六位校验 TOTP, 其余的数据库校验密码
	pl := len(password)
	if pl <= 7 {
		err = fmt.Errorf("incorrect username or password")
		return
	}
	pinCode := password[:pl-6]
	otp := password[pl-6:]

	// 用数据库校验密码
	u := &accountModel.User{
		Username: username,
		Password: pinCode,
	}
	userRight, err := u.Login()
	if err != nil && userRight == nil {
		loginAttempt.Times++
		loginAttempt.LastFailedAt = time.Now()
		loginAttemptsCache.Set(username, loginAttempt, cache.DefaultExpiration)

		global.Log.Debugf("radius cache: %s-after: %+v attempt %+v", username, loginAttempt.Times, loginAttempt.LastFailedAt)
		return
	}
	// 校验 totp
	if totpModel.CheckTotp(userRight.Totp.Secret, otp) {
		valid = true
		// 清除该用户的登录失败记录，因为验证成功了
		loginAttemptsCache.Delete(username)
		return
	}

	loginAttempt.Times++
	loginAttempt.LastFailedAt = time.Now()
	loginAttemptsCache.Set(username, loginAttempt, cache.DefaultExpiration)

	global.Log.Debugf("radius cache: %s-after: %+v attempt %+v", username, loginAttempt.Times, loginAttempt.LastFailedAt)
	return
}

func AuthHandler(w radius.ResponseWriter, r *radius.Request) {
	username := rfc2865.UserName_GetString(r.Packet)
	password := rfc2865.UserPassword_GetString(r.Packet)

	code := radius.CodeAccessReject

	if userValid, err := AuthRequest(username, password); err != nil {
		global.Log.Errorf("Error while performing auth for user %s: %s", username, err)
	} else if userValid {
		code = radius.CodeAccessAccept
	}
	global.Log.Infof("Writing %v to %v", code, r.RemoteAddr)
	err := w.Write(r.Response(code))
	if err != nil {
		global.Log.Errorf("Encountered error when responding to request: %s", err)
	}
}

func NewRadiusServer() *radius.PacketServer {
	server := &radius.PacketServer{
		Addr:         config.Conf.Radius.ListenAddr,
		Handler:      radius.HandlerFunc(AuthHandler),
		SecretSource: radius.StaticSecretSource([]byte(config.Conf.Radius.Secret)),
	}

	global.Log.Infof("New radius server on: %s", config.Conf.Radius.ListenAddr)
	return server
}

// Radius Server Usage
func Run() (err error) {
	radiusServer := NewRadiusServer()
	return radiusServer.ListenAndServe()
}
