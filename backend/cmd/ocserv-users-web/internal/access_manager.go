package internal

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// EnforceVpnAccess checks all sessions and disconnects those that violate access rules (IP + time).
func EnforceVpnAccess(cfg []VpnAccessRule) error {
	return enforceVpnAccess(cfg)
}

// EnforceVpnAccessTime checks sessions and handles those that violate time-range rules only.
func EnforceVpnAccessOnlyByTime(cfg []VpnAccessRule) error {
	return enforceVpnAccess(cfg, true)
}

// InitializeTimeLocation sets the global timezone for VPN access time checks.
// This must be called once during initialization to ensure consistent time checking across all rules.
func InitializeTimeLocation(timezoneName string) error {
	var loc *time.Location
	var err error

	if timezoneName == "" || timezoneName == "UTC" {
		loc = time.UTC
	} else {
		// Try to load the specified timezone
		loc, err = time.LoadLocation(timezoneName)
		if err != nil {
			return fmt.Errorf("failed to load timezone %q: %w", timezoneName, err)
		}
	}

	globalTimeLocation = loc
	log.Printf("[access] initialized timezone for time checks: %s (location: %s)", timezoneName, loc)
	return nil
}

// RunVpnAccessTimeEnforcer periodically enforces time-range rules only.
func RunVpnAccessTimeEnforcer(ctx context.Context, refresh time.Duration) {
	log.Printf("[access] 启动 VPN 访问时间控制器 (refresh=%s, timezone=%s)", refresh.String(), globalTimeLocation)
	ticker := time.NewTicker(refresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[access] 停止访问控制器")
			return
		case <-ticker.C:
			// Get a thread-safe copy of the rules
			rules := GetVpnAccessRules()
			if err := enforceVpnAccess(rules, true); err != nil {
				log.Printf("[access] 强制访问失败: %v", err)
				continue
			}
		}
	}
}

// enforceVpnAccess is the common enforcement logic for all access rules.
// checkIPWhitelist determines whether to also check IP whitelist.
// Improved error handling with partial failure tolerance.
func enforceVpnAccess(cfg []VpnAccessRule, onlyCheckTime ...bool) error {
	// Set default value: onlyCheckTime = false
	check := false
	if len(onlyCheckTime) > 0 {
		check = onlyCheckTime[0]
	}

	sessions, err := OcctlGetSessions()
	if err != nil {
		return fmt.Errorf("[access] 获取会话失败: %v", err)
	}

	now := time.Now()
	successCount := 0
	failureCount := 0

	for _, s := range sessions {
		allow, action, reason := checkSession(cfg, s, now, check)

		if allow {
			continue
		}

		if action == nil {
			log.Printf("[access] action null pointer: id=%d user=%s remote=%s", s.ID, s.Username, s.RemoteIP)
			failureCount++
			continue
		}

		// Rule matched but session not allowed
		switch *action {
		case VpnActionBlock:
			if err := OcctlDisconnectUserByID(fmt.Sprintf("%d", s.ID)); err != nil {
				log.Printf("[access] 断开会话失败 id=%d: %v", s.ID, err)
				failureCount++
				continue
			}
			// Record violation to database (non-blocking now)
			if err := RecordViolation(s.Username, s.RemoteIP, string(VpnActionBlock), reason); err != nil {
				log.Printf("[violation] failed to record violation: %v", err)
				// Don't fail the entire operation for logging failures
			}
			log.Printf("[access] 会话不符合规则，断开: id=%d user=%s remote=%s action=block reason=%s", s.ID, s.Username, s.RemoteIP, reason)
			successCount++
		case VpnActionLogOnly:
			// Record violation to database (non-blocking now)
			if err := RecordViolation(s.Username, s.RemoteIP, string(VpnActionLogOnly), reason); err != nil {
				log.Printf("[violation] failed to record violation: %v", err)
				// Don't fail the entire operation for logging failures
			}
			log.Printf("[access] 会话不符合规则，仅记录: id=%d user=%s remote=%s action=logonly reason=%s", s.ID, s.Username, s.RemoteIP, reason)
			successCount++
		}
	}

	// Log summary
	if failureCount > 0 && len(sessions) > 0 {
		log.Printf("[access] 访问校验完成: 总计=%d, 成功=%d, 失败=%d", len(sessions), successCount, failureCount)
	}

	return nil
}

// checkSession determines whether the given session is permitted according to the config.
// Policy: if any rule that matches the session's username exists and its predicates pass, session is allowed.
// If there are rules for the user but none of them permit the session, it is disallowed.
// onlyCheckTime determines whether to only check time range (true) or check IP+time (false, default).
// Returns (allow, action, reason) where action is the rule action (or nil if no rule matched the user).
func checkSession(cfg []VpnAccessRule, s Session, now time.Time, onlyCheckTime bool) (allow bool, action *VpnActionType, reason string) {
	u := strings.ToLower(s.Username)
	hasRule := false
	var lastReason string

	for _, r := range cfg {
		if !r.matchesUser(u) {
			continue
		}
		hasRule = true

		// check other predicates if not only checking time
		if !onlyCheckTime {
			ipOk := r.ipInWhitelist(s.RemoteIP)
			if !ipOk {
				action = &r.Action
				lastReason = fmt.Sprintf("IP %s not in whitelist", s.RemoteIP)
				// IP not in whitelist, mean this rule not matched, check next rule
				continue
			}
		}

		// Check time range
		timeOk := r.isWithinTimeRange(now)
		if timeOk {
			// allowed by this rule
			return true, nil, ""
		}
		action = &r.Action
		if r.TimeRange != nil {
			lastReason = fmt.Sprintf("outside allowed time range %s-%s", r.TimeRange.Start.String(), r.TimeRange.End.String())
		} else {
			lastReason = "time check failed"
		}
	}

	// no rule for user -> not managed by access control (allow)
	if !hasRule {
		return true, nil, ""
	}

	// had rules but none matched
	return false, action, lastReason
}
