// env GOOS=linux GOARCH=amd64 go build -o ocserv-users-linux-amd64
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
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
	ID           int      `json:"ID"`
	Username     string   `json:"Username"`
	Groupname    string   `json:"Groupname"`
	State        string   `json:"State"`
	Vhost        string   `json:"vhost"`
	Device       string   `json:"Device"`
	RemoteIP     string   `json:"Remote IP"`
	IPv4         string   `json:"IPv4"`
	PtPIPv4      string   `json:"P-t-P IPv4"`
	DNS          []string `json:"DNS"`
	Routes       []string `json:"Routes"`
	NoRoutes     []string `json:"No-routes"`
	UserAgent    string   `json:"User-Agent"`
	RX           string   `json:"RX"` // raw bytes as string
	TX           string   `json:"TX"`
	AverageRX    string   `json:"Average RX"`
	AverageTX    string   `json:"Average TX"`
	ConnectedAt  string   `json:"Connected at"`  // string time
	ConnectedFor string   `json:"_Connected at"` // duration string

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

// -------------------- 全局变量 --------------------

var (
	conn  = &nftables.Conn{}
	table = &nftables.Table{
		Name:   "vpn_filter",
		Family: nftables.TableFamilyINet,
	}
	policy = nftables.ChainPolicyDrop
	chain  = &nftables.Chain{
		Name:     "vpn_forward",
		Table:    table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookForward,
		Priority: nftables.ChainPriorityFilter,
		Policy:   &policy,
	}
	onlineUsers = make(map[string]string) // username -> IPv4
)

// -------------------- 工具函数 --------------------

// IP 点分转 4 字节
func ipToBytes(ip string) []byte {
	var b [4]byte
	fmt.Sscanf(ip, "%d.%d.%d.%d", &b[0], &b[1], &b[2], &b[3])
	return b[:]
}

// 16-bit 转 bytes
func uint16ToBytes(n uint16) []byte {
	return []byte{byte(n >> 8), byte(n & 0xff)}
}

// 将 IP 或 CIDR 转成 nftables Bitwise 匹配需要的 Addr + Mask
func parseIPOrCIDR(ipstr string) (addr []byte, mask []byte) {
	if !strings.Contains(ipstr, "/") {
		// 单 IP
		addr = ipToBytes(ipstr)
		mask = []byte{0xff, 0xff, 0xff, 0xff}
		return
	}
	ip, ipnet, err := net.ParseCIDR(ipstr)
	if err != nil {
		log.Printf("CIDR 解析失败: %s", ipstr)
		return nil, nil
	}
	addr = ip.To4()
	mask = ipnet.Mask
	return
}

// -------------------- nftables 操作 --------------------

// 初始化表、链和公共默认规则
func initNftables(publicRules []RuleConfig) {
	conn.AddTable(table)
	conn.AddChain(chain)
	conn.Flush()

	// 添加公共规则
	for i, r := range publicRules {
		userTag := fmt.Sprintf("public:%d", i)
		addNftRule(userTag, "0.0.0.0", r) // src ip "0.0.0.0" 表示所有源
	}
	conn.Flush()
}

// 增量添加规则
func addNftRule(userTag string, srcIP string, rule RuleConfig) {
	proto := 0
	switch strings.ToLower(rule.Protocol) {
	case "tcp":
		proto = unix.IPPROTO_TCP
	case "udp":
		proto = unix.IPPROTO_UDP
	case "icmp":
		proto = unix.IPPROTO_ICMP
	default:
		log.Printf("未知协议: %s", rule.Protocol)
		return
	}

	addr, mask := parseIPOrCIDR(srcIP)
	if addr == nil || mask == nil {
		return
	}

	exprs := []expr.Any{
		// saddr + CIDR
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseNetworkHeader,
			Offset:       12,
			Len:          4,
		},
		&expr.Bitwise{
			SourceRegister: 1,
			Mask:           mask,
			Xor:            []byte{0, 0, 0, 0},
			DestRegister:   1,
		},
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpEq,
			Data:     addr,
		},
		// L4 protocol
		&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
		&expr.Cmp{Register: 1, Op: expr.CmpOpEq, Data: []byte{byte(proto)}},
	}

	// TCP/UDP 指定端口
	if rule.Port != 0 && (proto == unix.IPPROTO_TCP || proto == unix.IPPROTO_UDP) {
		exprs = append(exprs,
			&expr.Payload{
				Base:         expr.PayloadBaseTransportHeader,
				Offset:       2, // dport
				Len:          2,
				DestRegister: 1,
			},
			&expr.Cmp{
				Register: 1,
				Op:       expr.CmpOpEq,
				Data:     uint16ToBytes(rule.Port),
			})
	}

	exprs = append(exprs, &expr.Verdict{Kind: expr.VerdictAccept})

	conn.AddRule(&nftables.Rule{
		Table:    table,
		Chain:    chain,
		Exprs:    exprs,
		UserData: []byte(userTag),
	})
}

// 删除用户规则
func deleteUserRules(userTag string) {
	rules, _ := conn.GetRules(table, chain)
	for _, r := range rules {
		if string(r.UserData) == userTag {
			conn.DelRule(r)
		}
	}
	conn.Flush()
}

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

	http.HandleFunc("/", indexHandler)

	log.Println("服务运行中： http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("启动失败:", err)
	}

	configPath := "rules.json"
	refreshInterval := 30 * time.Second

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("读取配置失败: %v", err)
	}

	// 初始化表链和公共规则
	initNftables(cfg.Public)
	log.Printf("初始化完成，公共规则已添加")

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
			currentUsers[s.Username] = s.IPv4
			if oldIP, ok := onlineUsers[s.Username]; !ok || oldIP != s.IPv4 {
				// 新用户或 IP 变更
				userRules := cfg.Users[s.Username]
				for i, r := range userRules {
					userTag := fmt.Sprintf("user:%s:%d", s.Username, i)
					addNftRule(userTag, s.IPv4, r)
				}
				conn.Flush()
				log.Printf("用户 %s 的规则已添加", s.Username)
			}
		}

		// 检查下线用户
		for username := range onlineUsers {
			if _, ok := currentUsers[username]; !ok {
				userRules := cfg.Users[username]
				for i := range userRules {
					userTag := fmt.Sprintf("user:%s:%d", username, i)
					deleteUserRules(userTag)
				}
			}
		}

		onlineUsers = currentUsers
		time.Sleep(refreshInterval)
	}
}
