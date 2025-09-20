// env GOOS=linux GOARCH=amd64 go build -o ocserv-users-linux-amd64
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// -------------------- 数据结构 --------------------

type RuleConfig struct {
	IP       string `json:"ip"`
	Protocol string `json:"protocol"`
	Port     uint16 `json:"port,omitempty"`
}

type FullConfig struct {
	Public []RuleConfig            `json:"public"`
	Users  map[string][]RuleConfig `json:"users"`
}

// Session 表示每个用户 VPN 会话
type Session struct {
	ID           int         `json:"ID"`
	Username     string      `json:"Username"`
	Groupname    string      `json:"Groupname"`
	State        string      `json:"State"`
	Vhost        string      `json:"vhost"`
	Device       string      `json:"Device"`
	RemoteIP     string      `json:"Remote IP"`
	IPv4         string      `json:"IPv4"`
	PtPIPv4      string      `json:"P-t-P IPv4"`
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

func toHumanSize(bytesStr string) string {
	n, err := strconv.ParseInt(bytesStr, 10, 64)
	if err != nil {
		return "?"
	}
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case n > GB:
		return fmt.Sprintf("%.2f GB", float64(n)/float64(GB))
	case n > MB:
		return fmt.Sprintf("%.2f MB", float64(n)/float64(MB))
	case n > KB:
		return fmt.Sprintf("%.2f KB", float64(n)/float64(KB))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func parseStringOrSlice(v interface{}) []string {
	switch vv := v.(type) {
	case string:
		if vv == "" {
			return []string{}
		}
		return []string{vv}
	case []interface{}:
		var out []string
		for _, x := range vv {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return []string{}
	}
}

var (
	onlineUsers = make(map[string]string) // username -> IPv4
)

// -------------------- nftables 操作 --------------------

// 通过 nft 命令检查表和链是否存在，如果不存在则创建
func ensureNatTableAndChain() error {
	// 检查 nat 表是否存在
	cmd := exec.Command("nft", "list", "table", "ip", "nat")
	if err := cmd.Run(); err != nil {
		// 表不存在，创建表
		cmd = exec.Command("nft", "add", "table", "ip", "nat")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create nat table: %v", err)
		}
	}

	// 检查 POSTROUTING 链是否存在
	cmd = exec.Command("nft", "list", "chain", "ip", "nat", "POSTROUTING")
	if err := cmd.Run(); err != nil {
		// 链不存在，创建链
		cmd = exec.Command("nft", "add", "chain", "ip", "nat", "POSTROUTING",
			"{", "type", "nat", "hook", "postrouting", "priority", "srcnat;", "policy", "accept;", "}")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create POSTROUTING chain: %v", err)
		}
	}

	// 检查是否已有 masquerade 规则
	cmd = exec.Command("nft", "list", "chain", "ip", "nat", "POSTROUTING")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list POSTROUTING chain: %v", err)
	}
	if !strings.Contains(string(output), "masquerade") {
		// 添加 masquerade 规则
		cmd = exec.Command("nft", "add", "rule", "ip", "nat", "POSTROUTING", "masquerade")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add masquerade rule: %v", err)
		}
		log.Println("已添加 nat POSTROUTING masquerade 规则")
	} else {
		log.Println("nat POSTROUTING masquerade 规则已存在")
	}
	return nil
}

func ensureFilterTableAndChain() error {
	// 检查 vpn_filter 表是否存在
	cmd := exec.Command("nft", "list", "table", "ip", "vpn_filter")
	if err := cmd.Run(); err != nil {
		// 表不存在，创建表
		cmd = exec.Command("nft", "add", "table", "ip", "vpn_filter")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create vpn_filter table: %v", err)
		}
	}

	// 检查 vpn_forward 链是否存在
	cmd = exec.Command("nft", "list", "chain", "ip", "vpn_filter", "vpn_forward")
	if err := cmd.Run(); err != nil {
		// 链不存在，创建链
		cmd = exec.Command("nft", "add", "chain", "ip", "vpn_filter", "vpn_forward",
			"{", "type", "filter", "hook", "forward", "priority", "filter;", "policy", "drop;", "}")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create vpn_forward chain: %v", err)
		}
	}
	return nil
}

// 初始化表、链和公共默认规则
func initNftables(publicRules []RuleConfig) {
	// 确保表/链存在
	if err := ensureNatTableAndChain(); err != nil {
		log.Fatalf("Failed to ensure table and chain: %v", err)
	}

	// 新增：确保 filter 表和链存在
	if err := ensureFilterTableAndChain(); err != nil {
		log.Fatalf("Failed to ensure filter table and chain: %v", err)
	}

	// 清除现有规则
	cmd := exec.Command("nft", "flush", "chain", "ip", "vpn_filter", "vpn_forward")
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to flush chain: %v", err)
	}

	// 添加已建立连接回包规则
	cmd = exec.Command("nft", "add", "rule", "ip", "vpn_filter", "vpn_forward",
		"position", "0",
		"ct", "state", "established,related",
		"accept", "comment", "\"allow return traffic\"")
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to add established/related rule: %v", err)
	}

	// 添加公共规则（源改为 0.0.0.0/0 表示任意源）
	for i, r := range publicRules {
		userTag := fmt.Sprintf("public:%d", i)
		addNftRule(userTag, "0.0.0.0/0", r)
	}
}

// 通过 nft 命令添加规则
func addNftRule(userTag string, srcIP string, rule RuleConfig) {
	proto := strings.ToLower(rule.Protocol)

	switch proto {
	case "tcp", "udp":
		var args []string
		if rule.Port != 0 {
			// 有端口时直接用协议和端口
			args = []string{
				"add", "rule", "ip", "vpn_filter", "vpn_forward",
				"position", "0",
				"ip", "saddr", srcIP,
				"ip", "daddr", rule.IP,
				proto, "dport", fmt.Sprintf("%d", rule.Port),
				"accept", "comment", fmt.Sprintf("\"%s\"", userTag),
			}
		} else {
			// 无端口时用 meta l4proto
			args = []string{
				"add", "rule", "ip", "vpn_filter", "vpn_forward",
				"position", "0",
				"ip", "saddr", srcIP,
				"ip", "daddr", rule.IP,
				"meta", "l4proto", proto,
				"accept", "comment", fmt.Sprintf("\"%s\"", userTag),
			}
		}
		cmd := exec.Command("nft", args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("规则 %s 添加失败: %v, output: %s", userTag, err, string(output))
			return
		}
		if rule.Port != 0 {
			log.Printf("规则 %s 添加成功: src=%s, dst=%s, proto=%s, port=%d",
				userTag, srcIP, rule.IP, rule.Protocol, rule.Port)
		} else {
			log.Printf("规则 %s 添加成功: src=%s, dst=%s, proto=%s (无端口)",
				userTag, srcIP, rule.IP, rule.Protocol)
		}
	case "icmp":
		cmd := exec.Command("nft", "add", "rule", "ip", "vpn_filter", "vpn_forward",
			"position", "0",
			"ip", "saddr", srcIP,
			"ip", "daddr", rule.IP,
			proto,
			"accept", "comment", fmt.Sprintf("\"%s\"", userTag))
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("规则 %s 添加失败: %v, output: %s", userTag, err, string(output))
			return
		}
		log.Printf("规则 %s 添加成功: src=%s, dst=%s, proto=%s",
			userTag, srcIP, rule.IP, rule.Protocol)
	default:
		log.Printf("未知协议: %s", rule.Protocol)
		return
	}
}

// 通过 nft 命令删除用户规则
func deleteNftRules(userTag string) {
	// 使用 nft list ruleset 并解析输出来找到带有特定注释的规则
	cmd := exec.Command("nft", "-a", "list", "chain", "ip", "vpn_filter", "vpn_forward")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("获取规则失败: %v", err)
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "# handle") && strings.Contains(line, userTag) {
			// 提取句柄编号
			handle := ""
			if idx := strings.LastIndex(line, "# handle "); idx != -1 {
				handle = strings.TrimSpace(line[idx+len("# handle "):])
			}

			if handle != "" {
				// 删除规则
				cmd := exec.Command("nft", "delete", "rule", "ip", "vpn_filter", "vpn_forward", "handle", handle)
				if err := cmd.Run(); err != nil {
					log.Printf("删除规则失败 %s: %v", userTag, err)
				} else {
					log.Printf("删除规则成功: %s", userTag)
				}
			}
		}
	}
}

// -------------------- occtl user 操作 --------------------

func getSessions() ([]Session, error) {
	cmd := exec.Command("occtl", "-j", "show", "users")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var sessions []Session
	if err := json.Unmarshal(output, &sessions); err != nil {
		return nil, err
	}

	for i := range sessions {
		sessions[i].RXHuman = toHumanSize(sessions[i].RX)
		sessions[i].TXHuman = toHumanSize(sessions[i].TX)
		// 兼容 Routes/NoRoutes 字段
		sessions[i].Routes = parseStringOrSlice(sessions[i].Routes)
		sessions[i].NoRoutes = parseStringOrSlice(sessions[i].NoRoutes)
	}

	return sessions, nil
}

// 读取 JSON 配置
func loadConfig(path string) (FullConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FullConfig{}, err
	}
	var cfg FullConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return FullConfig{}, err
	}
	return cfg, nil
}

// -------------------- Web handler --------------------

//go:embed static/index.html
var Static embed.FS

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(Static, "static/index.html"))

	sessions, err := getSessions()
	if err != nil {
		http.Error(w, "无法获取用户数据: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, sessions)
}

// -------------------- 主循环 --------------------
func main() {
	// 移除 addSimpleTcp22Rule 调用，因为我们现在使用命令行方式

	http.HandleFunc("/", indexHandler)

	go func() {
		log.Println("服务运行中： http://localhost:8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal("启动失败:", err)
		}
	}()

	configPath := "rules.json"
	refreshInterval := 30 * time.Second

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("读取配置失败: %v", err)
	}

	// 初始化表链和公共规则
	initNftables(cfg.Public)
	log.Printf("初始化完成，公共规则已添加")

	go func() {
		for {
			// 加载最新配置（可支持热更新）
			cfg, err = loadConfig(configPath)
			if err != nil {
				log.Printf("读取配置失败: %v", err)
				time.Sleep(refreshInterval)
				continue
			}

			sessions, err := getSessions()
			if err != nil {
				log.Printf("获取会话失败: %v", err)
				time.Sleep(refreshInterval)
				continue
			}

			currentUsers := make(map[string]string)
			for _, s := range sessions {
				log.Printf("Session: Username=%s, IPv4=%s", s.Username, s.IPv4)

				currentUsers[s.Username] = s.IPv4
				if oldIP, ok := onlineUsers[s.Username]; !ok || oldIP != s.IPv4 {
					// 新用户或 IP 变更
					userRules := cfg.Users[s.Username]
					for i, r := range userRules {
						log.Printf("userRules set: Username=%s, rno=%d, rule=%+v", s.Username, i, r)
						userTag := fmt.Sprintf("user:%s:%d", s.Username, i)
						// 如果 s.IPv4 为空，跳过
						if s.IPv4 == "" {
							log.Printf("警告: 用户 %s 的 IPv4 为空，跳过添加规则", s.Username)
							continue
						}
						addNftRule(userTag, s.IPv4, r)
					}
					log.Printf("用户 %s 的规则已处理", s.Username)
				}
			}

			// 检查下线用户
			for username := range onlineUsers {
				if _, ok := currentUsers[username]; !ok {
					userRules := cfg.Users[username]
					for i := range userRules {
						// 删除规则
						log.Printf("userRules unset: %s:%d", username, i)
						userTag := fmt.Sprintf("user:%s:%d", username, i)
						deleteNftRules(userTag)
					}
				}
			}

			onlineUsers = currentUsers
			time.Sleep(refreshInterval)
		}
	}()
	select {}
}
