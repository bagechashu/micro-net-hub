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
	"context"
	"fmt"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
	accountModel "micro-net-hub/internal/module/account/model"
	"micro-net-hub/internal/module/approval"
	totpModel "micro-net-hub/internal/module/totp/model"

	"github.com/patrickmn/go-cache"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type loginAttemptInfo struct {
	Times        int
	LastFailedAt time.Time
}

// checkApproval 人工审批门禁: 返回 nil 表示放行, 否则返回拒绝原因.
//
// 门禁在 TOTP 校验通过之后执行, 因此审批人不会被"密码错误的刷屏请求"打扰;
// 相应地, 审批相关的拒绝也不写入登录失败计数, 由 reject-cooldown-seconds 限流.
func checkApproval(ctx context.Context, user *accountModel.User, meta approval.Meta) error {
	if !approval.Enabled() {
		return nil
	}

	allow, err := approval.Gate(ctx, user, meta)
	if err != nil {
		global.Log.Warnf("人工审批未放行: username=%s, %v", user.Username, err)
		return err
	}
	if !allow {
		return fmt.Errorf("登录未获人工审批放行, username=%s", user.Username)
	}
	return nil
}

// 初始化一个缓存实例，设置过期时间为 banDurationMinute 分钟
var banDurationMinute float64 = 5
var loginAttemptsCache = cache.New(time.Duration(banDurationMinute)*time.Minute, time.Duration(banDurationMinute)*time.Minute)

// AuthRequest 校验用户名密码与 TOTP, 通过后按需进入人工审批门禁.
//
// 注意: 人工审批的拒绝与等待超时都不写入 loginAttemptsCache, 否则用户重试几次
// 就会触发 5 分钟锁定, 反而放大故障面; 重复申请的限流由 reject-cooldown-seconds 承担.
func AuthRequest(ctx context.Context, username string, password string, meta approval.Meta) (valid bool, err error) {
	var loginAttempt = &loginAttemptInfo{}
	attempt, found := loginAttemptsCache.Get(username)
	if found {
		loginAttempt = attempt.(*loginAttemptInfo)
		global.Log.Debugf("radius cache: %s-before: %+v attempt %+v", username, loginAttempt.Times, loginAttempt.LastFailedAt)

		// 如果最后一次失败尝试距离现在不足10分钟，则返回错误并禁止登录
		if loginAttempt.Times > config.Conf.Radius.FailTimesBeforeBlock5min && time.Since(loginAttempt.LastFailedAt).Minutes() < banDurationMinute {
			return false, fmt.Errorf("登录失败次数过多，账户已被锁定5分钟, username=%s", username)
		}
	}

	// password 后六位校验 TOTP, 其余的数据库校验密码
	pl := len(password)
	if pl <= 7 {
		return false, fmt.Errorf("incorrect username or password, username=%s", username)
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
		// 登录失败次数记录
		loginAttempt.Times++
		loginAttempt.LastFailedAt = time.Now()
		loginAttemptsCache.Set(username, loginAttempt, cache.DefaultExpiration)

		// global.Log.Debugf("radius cache: %s-after: %+v attempt %+v", username, loginAttempt.Times, loginAttempt.LastFailedAt)
		return false, fmt.Errorf("incorrect username or password, username=%s", username)
	}
	// 禁用是 BindDN 的角色登录
	if userRight.CheckAdminDN() {
		return false, fmt.Errorf("AdminDN 禁止登录 使用 radius, username=%s", username)
	}
	// 禁用是 BindDN 的角色登录
	if userRight.CheckBindDNRole() {
		return false, fmt.Errorf("用户为 BindDN Role, 禁止登录, username=%s", username)
	}
	// 校验 totp
	if totpModel.CheckTotp(userRight.Totp.Secret, otp) {
		// TOTP 通过后才进入人工审批: 避免未通过身份校验的请求消耗审批人注意力
		if err := checkApproval(ctx, userRight, meta); err != nil {
			return false, err
		}

		// 验证成功了,清除该用户的登录失败记录，
		loginAttemptsCache.Delete(username)
		return true, nil
	}

	// Totp验证失败, 记录失败次数
	loginAttempt.Times++
	loginAttempt.LastFailedAt = time.Now()
	loginAttemptsCache.Set(username, loginAttempt, cache.DefaultExpiration)

	// global.Log.Debugf("radius cache: %s-after: %+v attempt %+v", username, loginAttempt.Times, loginAttempt.LastFailedAt)
	return false, fmt.Errorf("totp 验证失败, username=%s", username)
}

func AuthHandler(w radius.ResponseWriter, r *radius.Request) {
	username := rfc2865.UserName_GetString(r.Packet)
	password := rfc2865.UserPassword_GetString(r.Packet)

	// 来源与设备信息仅用于审批单留痕, 取不到不影响认证流程
	remoteAddr := ""
	if r.RemoteAddr != nil {
		remoteAddr = r.RemoteAddr.String()
	}
	nasIP := ""
	if ip := rfc2865.NASIPAddress_Get(r.Packet); ip != nil {
		nasIP = ip.String()
	}
	// 把 RADIUS 协议属性映射为审批模块的通用认证上下文
	meta := approval.Meta{
		approval.MetaKeySourceAddr: remoteAddr,
		approval.MetaKeySourceID:   rfc2865.NASIdentifier_GetString(r.Packet),
		approval.MetaKeySourceIP:   nasIP,
	}

	code := radius.CodeAccessReject

	if userValid, err := AuthRequest(r.Context(), username, password, meta); err != nil {
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
