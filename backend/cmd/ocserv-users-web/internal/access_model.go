package internal

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sigs.k8s.io/yaml"
)

var (
	// Global_VpnAccessRules 全局 VPN 访问控制规则
	Global_VpnAccessRules *VpnAccessConfig
)

// TimeRange defines a daily time range in HH:MM format, e.g. {"start":"08:00","end":"18:00"}
// It matches time-of-day and supports ranges that wrap over midnight (e.g., 22:00-06:00).
type TimeRange struct {
	Start *TimeOfDay `json:"start,omitempty"`
	End   *TimeOfDay `json:"end,omitempty"`
}

// TimeOfDay represents a time-of-day in 24-hour HH:MM format (hours 0-23).
type TimeOfDay struct {
	Hour   int `json:"-"`
	Minute int `json:"-"`
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

// UnmarshalJSON implements json.Unmarshaler
func (t *TimeOfDay) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("TimeOfDay should be string: %w", err)
	}
	return t.parse(s)
}

// MarshalJSON implements json.Marshaler
func (t TimeOfDay) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalText for yaml
func (t *TimeOfDay) UnmarshalText(b []byte) error {
	return t.parse(string(b))
}

// UnmarshalYAML for sigs.k8s.io/yaml
func (t *TimeOfDay) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	return t.parse(s)
}

// Contains reports whether the provided time (local) falls into the time range.
func (tr TimeRange) Contains(now time.Time) (bool, error) {
	if tr.Start == nil || tr.End == nil {
		return true, nil
	}
	y := now.Year()
	m := now.Month()
	d := now.Day()
	startT := time.Date(y, m, d, tr.Start.Hour, tr.Start.Minute, 0, 0, now.Location())
	endT := time.Date(y, m, d, tr.End.Hour, tr.End.Minute, 0, 0, now.Location())
	if !startT.Before(endT) { // wraps over midnight
		// allowed if now >= start (same day) or now <= end (next day)
		if now.Equal(startT) || now.After(startT) {
			return true, nil
		}
		if now.Before(endT) || now.Equal(endT) {
			return true, nil
		}
		return false, nil
	}

	// log.Printf("[access] 检查时间范围: now=%s start=%s end=%s", now.Format("15:04"), startT.Format("15:04"), endT.Format("15:04"))
	// normal case
	if (now.Equal(startT) || now.After(startT)) && (now.Equal(endT) || now.Before(endT)) {
		return true, nil
	}
	return false, nil
}

// VpnAccessRule describes access constraints for a set of users.
type VpnAccessRule struct {
	Users             []string   `json:"users"`
	RemoteIPWhiteList []string   `json:"remote_ip_whitelist,omitempty"`
	TimeRange         *TimeRange `json:"time_range,omitempty"` // e.g. "08:00-18:00"
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

// ipInWhitelist checks whether remote ip matches any whitelist entry (CIDR or single IP)
func (r VpnAccessRule) ipInWhitelist(remoteIP string) bool {
	if len(r.RemoteIPWhiteList) == 0 {
		return true // no whitelist means allow any IP
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

	for _, entry := range r.RemoteIPWhiteList {
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

// isWithinTimeRange checks time constraint (or allows if no time range specified)
func (r VpnAccessRule) isWithinTimeRange(now time.Time) (bool, error) {
	if r.TimeRange == nil {
		return true, nil
	}
	return r.TimeRange.Contains(now)
}

// VpnAccessConfig holds multiple VpnAccessRule entries
type VpnAccessConfig struct {
	VpnAccessRules []VpnAccessRule `json:"vpn_access_rules,omitempty"`
}

// LoadVpnAccessConfig loads the file (json or yaml) from path
func LoadVpnAccessConfig(path string) (*VpnAccessConfig, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg VpnAccessConfig

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse yaml config %q: %w", path, err)
		}
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse json config %q: %w", path, err)
		}
	default:
		// Try json first, then yaml
		if err := json.Unmarshal(data, &cfg); err == nil {
			break
		}
		if err2 := yaml.Unmarshal(data, &cfg); err2 == nil {
			break
		} else {
			return nil, fmt.Errorf("failed to parse config %q as json or yaml: json: %v; yaml: %v", path, err, err2)
		}
	}

	return &cfg, nil
}
