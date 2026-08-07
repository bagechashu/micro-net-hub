package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// -------------------- 全局状态 --------------------

var (
	globalOnlineOcUsers   = make(map[string][]string) // username -> []IPv4
	globalOnlineOcUsersMu sync.Mutex
)

// 支持多设备，tag = username:ip:ruleIndex
func checkOcSessions() (addedSessions, removedSessions map[string][]string, err error) {
	addedSessions = make(map[string][]string)
	removedSessions = make(map[string][]string)

	sessions, err := OcctlGetSessions()
	if err != nil {
		log.Printf("[nft] 获取会话失败: %v", err)
		return nil, nil, err
	}

	// 当前在线用户映射 username -> []IPv4
	currentUsers := buildCurrentOcUsersFromSessions(sessions)

	globalOnlineOcUsersMu.Lock()
	defer globalOnlineOcUsersMu.Unlock()

	// 处理现有在线用户的 IP 变化
	for username, ips := range currentUsers {
		oldIPs := globalOnlineOcUsers[username]

		added, removed := diffIPs(oldIPs, ips)
		if len(added) > 0 {
			log.Printf("[diff] 用户 %s 上线 IP: %v", username, added)
			addedSessions[username] = append(addedSessions[username], added...)
		}
		if len(removed) > 0 {
			log.Printf("[diff] 用户 %s 下线 IP: %v", username, removed)
			removedSessions[username] = append(removedSessions[username], removed...)
		}
	}

	// 完全下线用户
	for username, offlineIps := range globalOnlineOcUsers {
		if _, ok := currentUsers[username]; !ok {
			log.Printf("[diff] 用户 %s 完全下线，IP: %v", username, offlineIps)
			removedSessions[username] = append(removedSessions[username], offlineIps...)
		}
	}

	// 更新全局在线用户映射
	globalOnlineOcUsers = currentUsers

	return addedSessions, removedSessions, nil
}

// OfflineAllSessions 清空所有在线用户状态（用于重置）
func offlineAllSessions() {
	globalOnlineOcUsersMu.Lock()
	defer globalOnlineOcUsersMu.Unlock()

	globalOnlineOcUsers = make(map[string][]string)
}

// buildCurrentOcUsersFromSessions 从会话数据构建当前在线用户映射
func buildCurrentOcUsersFromSessions(sessions []Session) map[string][]string {
	current := make(map[string][]string)
	for _, s := range sessions {
		if s.Username == "" || s.IPv4 == "" {
			continue
		}
		current[s.Username] = append(current[s.Username], s.IPv4)
	}
	return current
}

// diffIPs 对比两个 IP 列表，返回新增和移除的 IP
func diffIPs(oldIPs, newIPs []string) (added, removed []string) {
	oldMap := make(map[string]struct{}, len(oldIPs))
	newMap := make(map[string]struct{}, len(newIPs))

	for _, ip := range oldIPs {
		oldMap[ip] = struct{}{}
	}
	for _, ip := range newIPs {
		newMap[ip] = struct{}{}
	}

	for ip := range newMap {
		if _, ok := oldMap[ip]; !ok {
			added = append(added, ip)
		}
	}
	for ip := range oldMap {
		if _, ok := newMap[ip]; !ok {
			removed = append(removed, ip)
		}
	}

	return
}

// Session represents an ocserv user session.
type Session struct {
	ID        int    `json:"ID"`
	Username  string `json:"Username"`
	Groupname string `json:"Groupname"`
	State     string `json:"State"`
	// Vhost        string      `json:"vhost"`
	Device   string `json:"Device"`
	RemoteIP string `json:"Remote IP"`
	IPv4     string `json:"IPv4"`
	// PtPIPv4      string      `json:"P-t-P IPv4"`
	DNS          []string    `json:"DNS"`
	Routes       interface{} `json:"Routes"`
	NoRoutes     interface{} `json:"No-routes"`
	UserAgent    string      `json:"User-Agent"`
	RX           string      `json:"RX"` // raw bytes as string
	TX           string      `json:"TX"`
	AverageRX    string      `json:"Average RX"`
	AverageTX    string      `json:"Average TX"`
	ConnectedAt  string      `json:"Connected at"`  // string time
	ConnectedFor string      `json:"_Connected at"` // duration string

	RXHuman string `json:"RXHuman"`
	TXHuman string `json:"TXHuman"`
}

// ValidateSessionID validates that a session ID contains only numeric characters
func ValidateSessionID(id string) error {
	if id == "" {
		return fmt.Errorf("session id cannot be empty")
	}
	// Session ID must be numeric only
	if !regexp.MustCompile(`^\d+$`).MatchString(id) {
		return fmt.Errorf("invalid session id: must contain only digits, got %q", id)
	}
	return nil
}

func OcctlDisconnectUserByID(id string) error {
	// Validate input before executing command
	if err := ValidateSessionID(id); err != nil {
		return fmt.Errorf("invalid session id: %w", err)
	}
	cmd := exec.Command("occtl", "disconnect", "id", id)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disconnect session %s: %w", id, err)
	}
	return nil
}

// OcctlGetSessions 使用 occtl 获取当前会话列表
func OcctlGetSessions() ([]Session, error) {
	time.Sleep(500 * time.Millisecond) // 等待 ocserv 稳定
	cmd := exec.Command("occtl", "-j", "show", "users")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var sessions []Session
	if err := json.Unmarshal(output, &sessions); err != nil {
		return nil, err
	}

	// 数据清洗
	filtered := make([]Session, 0, len(sessions))
	for i := range sessions {
		s := &sessions[i]

		// 过滤掉 username=none 的会话
		if s.Username == "" || s.Username == "(none)" {
			continue
		}

		// s.Username 全部转成小写, 因为 Ocserv 登录的用户不区分大小写
		s.Username = strings.ToLower(s.Username)

		s.RXHuman = toHumanSize(s.RX)
		s.TXHuman = toHumanSize(s.TX)
		s.Routes = parseStringOrSlice(s.Routes)
		s.NoRoutes = parseStringOrSlice(s.NoRoutes)

		filtered = append(filtered, *s)
	}

	return filtered, nil
}

func toHumanSize(bytesStr string) string {
	n, err := strconv.ParseInt(bytesStr, 10, 64)
	if err != nil {
		return "?"
	}
	const KB, MB, GB = 1024, 1024 * 1024, 1024 * 1024 * 1024
	switch {
	case n >= GB:
		return fmt.Sprintf("%.2f GB", float64(n)/float64(GB))
	case n >= MB:
		return fmt.Sprintf("%.2f MB", float64(n)/float64(MB))
	case n >= KB:
		return fmt.Sprintf("%.2f KB", float64(n)/float64(KB))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func parseStringOrSlice(v interface{}) []string {
	switch vv := v.(type) {
	case string:
		if vv == "" {
			return nil
		}
		return []string{vv}
	case []interface{}:
		out := make([]string, 0, len(vv))
		for _, x := range vv {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
