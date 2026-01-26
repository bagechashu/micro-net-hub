package internal

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	// globalVpnAccessRules 全局 VPN 访问控制规则，受 globalVpnAccessRulesMu 保护
	globalVpnAccessRules []VpnAccessRule
	// globalVpnAccessRulesMu 保护全局 VPN 规则的并发访问
	globalVpnAccessRulesMu sync.RWMutex
	// globalTimeLocation 用于时间检查的全局时区，由 main 初始化
	globalTimeLocation *time.Location = time.UTC
)

// GetVpnAccessRules returns a thread-safe copy of the global VPN access rules
func GetVpnAccessRules() []VpnAccessRule {
	globalVpnAccessRulesMu.RLock()
	defer globalVpnAccessRulesMu.RUnlock()
	// Return a copy to prevent external modifications
	if globalVpnAccessRules == nil {
		return nil
	}
	rules := make([]VpnAccessRule, len(globalVpnAccessRules))
	copy(rules, globalVpnAccessRules)
	return rules
}

// UpdateVpnAccessRules safely updates the global VPN access rules
func UpdateVpnAccessRules(rules []VpnAccessRule) {
	globalVpnAccessRulesMu.Lock()
	defer globalVpnAccessRulesMu.Unlock()
	globalVpnAccessRules = rules
}

// TimeRange defines a daily time range in HH:MM format, e.g. {"start":"08:00","end":"18:00"}
// It matches time-of-day and supports ranges that wrap over midnight (e.g., 22:00-06:00).
type TimeRange struct {
	Start *TimeOfDay `json:"start,omitempty" yaml:"start,omitempty"`
	End   *TimeOfDay `json:"end,omitempty" yaml:"end,omitempty"`
}

// TimeOfDay represents a time-of-day in 24-hour HH:MM format (hours 0-23).
type TimeOfDay struct {
	Hour   int `json:"-" yaml:"-"`
	Minute int `json:"-" yaml:"-"`
}

// String returns HH:MM
func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

func (t *TimeOfDay) parse(s string) error {
	if s == "" {
		return fmt.Errorf("empty time")
	}
	parsed, err := time.Parse("15:04", s)
	if err != nil {
		return fmt.Errorf("invalid time %q: %w", s, err)
	}
	if parsed.Hour() < 0 || parsed.Hour() > 23 || parsed.Minute() < 0 || parsed.Minute() > 59 {
		return fmt.Errorf("time out of range: %q", s)
	}
	t.Hour = parsed.Hour()
	t.Minute = parsed.Minute()
	return nil
}

// MarshalJSON implements json.Marshaler
func (t TimeOfDay) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// MarshalText implements encoding.TextMarshaler for YAML
func (t TimeOfDay) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// MarshalYAML implements yaml.Marshaler for gopkg.in/yaml.v3
func (t TimeOfDay) MarshalYAML() (interface{}, error) {
	return t.String(), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (t *TimeOfDay) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("TimeOfDay should be string: %w", err)
	}
	return t.parse(s)
}

// UnmarshalText for yaml
func (t *TimeOfDay) UnmarshalText(b []byte) error {
	return t.parse(string(b))
}

// UnmarshalYAML for gopkg.in/yaml.v3
func (t *TimeOfDay) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	return t.parse(s)
}

// Contains reports whether the provided time (local) falls into the time range.
// Uses the global timezone setting (globalTimeLocation) for consistency.
func (tr TimeRange) Contains(now time.Time) bool {
	if tr.Start == nil || tr.End == nil {
		return true
	}

	// Convert to the designated timezone for consistent time checking
	now = now.In(globalTimeLocation)

	y := now.Year()
	m := now.Month()
	d := now.Day()
	startT := time.Date(y, m, d, tr.Start.Hour, tr.Start.Minute, 0, 0, globalTimeLocation)
	endT := time.Date(y, m, d, tr.End.Hour, tr.End.Minute, 0, 0, globalTimeLocation)

	if !startT.Before(endT) { // wraps over midnight
		// allowed if now >= start (same day) or now <= end (next day)
		if now.Equal(startT) || now.After(startT) {
			return true
		}
		if now.Before(endT) || now.Equal(endT) {
			return true
		}
		return false
	}

	// normal case: start < end on same day
	if (now.Equal(startT) || now.After(startT)) && (now.Equal(endT) || now.Before(endT)) {
		return true
	}
	return false
}

// VpnAccessRule describes access constraints for a set of users.
type VpnAccessRule struct {
	Users             []string      `json:"users,omitempty" yaml:"users,omitempty"`
	RemoteIPs         []string      `json:"remote_ips,omitempty" yaml:"remote_ips,omitempty"`
	TimeRange         *TimeRange    `json:"time_range,omitempty" yaml:"time_range,omitempty"`                   // e.g. "08:00-18:00"
	ActionOnViolation VpnActionType `json:"action_on_violation,omitempty" yaml:"action_on_violation,omitempty"` // "block" or "logonly" - action when user matches but violates constraints
}

type VpnActionType string

const (
	VpnActionBlock   VpnActionType = "block"
	VpnActionLogOnly VpnActionType = "logonly"
)

func (t VpnActionType) Valid() bool {
	switch t {
	case VpnActionBlock, VpnActionLogOnly:
		return true
	default:
		return false
	}
}

// matchesUser checks whether username is listed in rule.Users (case-insensitive)
func (r VpnAccessRule) matchesUser(username string) bool {
	u := strings.ToLower(username)
	for _, ru := range r.Users {
		if strings.ToLower(strings.TrimSpace(ru)) == u {
			return true
		}
	}
	return false
}

// isInRemoteIpsList checks whether remote ip matches any allowed IP entry (CIDR or single IP)
func (r VpnAccessRule) isInRemoteIps(remoteIP string) bool {
	if len(r.RemoteIPs) == 0 {
		return false // empty list means no IP is allowed
	}
	if remoteIP == "" {
		return false
	}
	ip := net.ParseIP(remoteIP)
	if ip == nil {
		// Sometimes remote IP may contain port (unlikely), try to strip
		if idx := strings.LastIndex(remoteIP, ":"); idx != -1 {
			ip = net.ParseIP(remoteIP[:idx])
		}
		if ip == nil {
			return false
		}
	}

	for _, entry := range r.RemoteIPs {
		// log.Printf("[access] 检查远程 IP %s 是否匹配白名单条目 %s", remoteIP, entry)
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// Try CIDR
		if _, ipnet, err := net.ParseCIDR(entry); err == nil {
			if ipnet.Contains(ip) {
				return true
			}
			continue
		}
		// Try single IP
		if other := net.ParseIP(entry); other != nil {
			if other.Equal(ip) {
				return true
			}
		}
	}
	return false
}

// isInTimeRange checks time constraint (or allows if no time range specified)
func (r VpnAccessRule) isInTimeRange(now time.Time) bool {
	if r.TimeRange == nil {
		return true
	}
	return r.TimeRange.Contains(now)
}
