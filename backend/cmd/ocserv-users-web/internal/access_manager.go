package internal

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// EnforceVpnAccess checks all sessions and disconnects those that violate access rules.
func EnforceVpnAccess(cfg []VpnAccessRule) error {
	sessions, err := OcctlGetSessions()
	if err != nil {
		return fmt.Errorf("[access] 获取会话失败: %v", err)
	}

	now := time.Now()
	for _, s := range sessions {
		allowed, err := checkSession(cfg, s, now)
		if err != nil {
			log.Printf("[access] 检查会话 %d (%s) 时出错: %v", s.ID, s.Username, err)
			continue
		}
		if !allowed {
			log.Printf("[access] 会话不符合规则，断开: id=%d user=%s remote=%s", s.ID, s.Username, s.RemoteIP)
			if err := OcctlDisconnectUserByID(fmt.Sprintf("%d", s.ID)); err != nil {
				log.Printf("[access] 断开会话失败 id=%d: %v", s.ID, err)
			}
		}
	}
	return nil
}

// EnforceVpnAccessTimeOnly checks sessions and disconnects those that violate time-range rules only.
func EnforceVpnAccessTime(cfg []VpnAccessRule) error {
	sessions, err := OcctlGetSessions()
	if err != nil {
		return fmt.Errorf("[access] 获取会话失败: %v", err)
	}

	now := time.Now()
	for _, s := range sessions {
		allowed, err := checkSessionTime(cfg, s, now)
		if err != nil {
			log.Printf("[access] 检查会话 %d (%s) 时间出错: %v", s.ID, s.Username, err)
			continue
		}
		if !allowed {
			log.Printf("[access] 会话超出允许时间范围，断开: id=%d user=%s remote=%s", s.ID, s.Username, s.RemoteIP)
			if err := OcctlDisconnectUserByID(fmt.Sprintf("%d", s.ID)); err != nil {
				log.Printf("[access] 断开会话失败 id=%d: %v", s.ID, err)
			}
		}
	}
	return nil
}

// RunVpnAccessTimeEnforcer periodically enforces time-range rules only.
func RunVpnAccessTimeEnforcer(ctx context.Context, refresh time.Duration) {
	log.Printf("[access] 启动 VPN 访问时间控制器 (refresh=%s) — **使用服务器时区**", refresh.String())
	ticker := time.NewTicker(refresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[access] 停止访问控制器")
			return
		case <-ticker.C:
			if err := EnforceVpnAccessTime(Global_VpnAccessRules); err != nil {
				log.Printf("[access] 强制访问失败: %v", err)
				continue
			}
		}
	}
}

// checkSession determines whether the given session is permitted according to the config.
// Policy: if any rule that matches the session's username exists and its predicates (ip/time) pass, session is allowed.
// If there are rules for the user but none of them permit the session, it is disallowed.
func checkSession(cfg []VpnAccessRule, s Session, now time.Time) (allow bool, err error) {
	u := strings.ToLower(s.Username)
	hasRule := false
	for _, r := range cfg {
		if !r.matchesUser(u) {
			continue
		}
		hasRule = true
		ipOk := r.ipInWhitelist(s.RemoteIP)
		timeOk, err := r.isWithinTimeRange(now)
		if err != nil {
			return false, err
		}
		if ipOk && timeOk {
			// allowed by this rule
			return true, nil
		}
	}
	if !hasRule {
		// no rule for user -> not managed by access control (allow)
		return true, nil
	}
	// had rules but none matched
	return false, nil
}

// checkSessionTime determines whether the given session is permitted according to the config.
// Policy: if any rule that matches the session's username exists and its predicates (time) pass, session is allowed.
// If there are rules for the user but none of them permit the session, it is disallowed.
func checkSessionTime(cfg []VpnAccessRule, s Session, now time.Time) (allow bool, err error) {
	u := strings.ToLower(s.Username)
	hasRule := false
	for _, r := range cfg {
		if !r.matchesUser(u) {
			continue
		}
		hasRule = true
		timeOk, err := r.isWithinTimeRange(now)
		if err != nil {
			return false, err
		}
		if timeOk {
			// allowed by this rule
			return true, nil
		}
	}
	if !hasRule {
		// no rule for user -> not managed by access control (allow)
		return true, nil
	}
	// had rules but none matched
	return false, nil
}
