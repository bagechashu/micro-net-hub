package internal

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// -------------------- Session Manager --------------------
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

	RXHuman string
	TXHuman string
}

// -------------------- 工具函数 --------------------

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

// GetSessions 使用 occtl 获取当前会话列表
func GetSessions() ([]Session, error) {
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
