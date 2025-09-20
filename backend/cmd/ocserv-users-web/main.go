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

// -------------------- 全局变量 --------------------

var (
	conn  = &nftables.Conn{} // FIX: 统一使用一个全局 conn
	table = &nftables.Table{
		Name:   "vpn_filter",
		Family: nftables.TableFamilyIPv4,
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

// ipToBytes: 使用 net.ParseIP 更稳健（返回 4 字节 IPv4）
func ipToBytes(ip string) []byte {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return nil
	}
	ip4 := parsed.To4()
	if ip4 == nil {
		return nil
	}
	return ip4
}

// 16-bit 转 bytes (big-endian)
func uint16ToBytes(n uint16) []byte {
	return []byte{byte(n >> 8), byte(n & 0xff)}
}

// 将 IP 或 CIDR 转成 nftables Bitwise 匹配需要的 Addr + Mask
func parseIPOrCIDR(ipstr string) (addr []byte, mask []byte) {
	// 如果用户写 0.0.0.0 并意图表示任意源，尽量把它当作 /0
	if ipstr == "0.0.0.0" {
		ipstr = "0.0.0.0/0"
	}
	if !strings.Contains(ipstr, "/") {
		// 单 IP
		ip4 := ipToBytes(ipstr)
		if ip4 == nil {
			log.Printf("IP 解析失败: %s", ipstr)
			return nil, nil
		}
		addr = ip4
		mask = []byte{0xff, 0xff, 0xff, 0xff}
		return
	}
	ip, ipnet, err := net.ParseCIDR(ipstr)
	if err != nil {
		log.Printf("CIDR 解析失败: %s (%v)", ipstr, err)
		return nil, nil
	}
	ip4 := ip.To4()
	if ip4 == nil {
		log.Printf("非 IPv4 地址: %s", ipstr)
		return nil, nil
	}
	addr = ip4
	mask = ipnet.Mask
	return
}

// -------------------- nftables 操作 --------------------

// getTableAndChain: 使用全局 conn，若不存在则创建后 Flush 并再次查询返回真实指针
func getTableAndChain() (*nftables.Table, *nftables.Chain, *nftables.Conn) {
	// 使用全局 conn
	// 查找表
	tables, err := conn.ListTables()
	if err != nil {
		log.Fatalf("ListTables 错误: %v", err)
	}
	var tbl *nftables.Table
	for _, t := range tables {
		if t.Name == table.Name && t.Family == table.Family {
			tbl = t
			break
		}
	}

	// 如果表不存在，创建
	if tbl == nil {
		log.Printf("nft 表 %s 不存在，创建它", table.Name)
		conn.AddTable(table)
		if err := conn.Flush(); err != nil {
			log.Fatalf("创建表并 Flush 失败: %v", err)
		}
		// 重新列出以获取系统中的对象指针
		tables, err = conn.ListTables()
		if err != nil {
			log.Fatalf("ListTables 失败: %v", err)
		}
		for _, t := range tables {
			if t.Name == table.Name && t.Family == table.Family {
				tbl = t
				break
			}
		}
		if tbl == nil {
			log.Fatalf("创建表后仍未找到 %s", table.Name)
		}
	}

	// 查找链
	chains, err := conn.ListChains()
	if err != nil {
		log.Fatalf("ListChains 错误: %v", err)
	}
	var chn *nftables.Chain
	for _, c := range chains {
		if c.Name == chain.Name && c.Table.Name == tbl.Name {
			chn = c
			break
		}
	}

	// 如果链不存在，创建
	if chn == nil {
		log.Printf("nft 链 %s 不存在，创建它", chain.Name)
		// 保证 chain.Table 指向正确的表
		chain.Table = tbl
		conn.AddChain(chain)
		if err := conn.Flush(); err != nil {
			log.Fatalf("创建链并 Flush 失败: %v", err)
		}
		// 重新列出获取指针
		chains, err = conn.ListChains()
		if err != nil {
			log.Fatalf("ListChains 失败: %v", err)
		}
		for _, c := range chains {
			if c.Name == chain.Name && c.Table.Name == tbl.Name {
				chn = c
				break
			}
		}
		if chn == nil {
			log.Fatalf("创建链后仍未找到 %s", chain.Name)
		}
	}

	return tbl, chn, conn
}

// 初始化表、链和公共默认规则
func initNftables(publicRules []RuleConfig) {
	// 确保表/链存在
	_, _, _ = getTableAndChain()

	// 添加公共规则（源改为 0.0.0.0/0 表示任意源）
	for i, r := range publicRules {
		userTag := fmt.Sprintf("public:%d", i)
		addNftRule(userTag, "0.0.0.0/0", r) // FIX: 使用 /0 表示任意源
	}
}

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

	dstAddr, dstMask := parseIPOrCIDR(rule.IP)
	if dstAddr == nil || dstMask == nil {
		log.Printf("目的地址无效: %s", rule.IP)
		return
	}

	var exprs []expr.Any

	// 只在非任意源地址时添加 src 匹配
	if srcIP != "0.0.0.0/0" {
		srcAddr, srcMask := parseIPOrCIDR(srcIP)
		if srcAddr != nil && srcMask != nil {
			exprs = append(exprs,
				&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 12, Len: 4},
				&expr.Bitwise{SourceRegister: 1, Mask: srcMask, Xor: []byte{0, 0, 0, 0}, DestRegister: 1},
				&expr.Cmp{Register: 1, Op: expr.CmpOpEq, Data: srcAddr},
			)
		}
	}

	// dst 匹配
	exprs = append(exprs,
		&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4},
		&expr.Bitwise{SourceRegister: 1, Mask: dstMask, Xor: []byte{0, 0, 0, 0}, DestRegister: 1},
		&expr.Cmp{Register: 1, Op: expr.CmpOpEq, Data: dstAddr},
	)

	// 协议匹配
	exprs = append(exprs,
		&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
		&expr.Cmp{Register: 1, Op: expr.CmpOpEq, Data: []byte{byte(proto)}},
	)

	// TCP/UDP 且端口不为 0 时才匹配 dport
	if rule.Port != 0 && (proto == unix.IPPROTO_TCP || proto == unix.IPPROTO_UDP) {
		exprs = append(exprs,
			&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseTransportHeader, Offset: 2, Len: 2},
			&expr.Cmp{Register: 1, Op: expr.CmpOpEq, Data: uint16ToBytes(rule.Port)},
		)
	}

	// accept
	exprs = append(exprs, &expr.Verdict{Kind: expr.VerdictAccept})

	tbl, chn, c := getTableAndChain()
	if c == nil || tbl == nil || chn == nil {
		log.Printf("table/chain 未就绪")
		return
	}

	r := &nftables.Rule{
		Table:    tbl,
		Chain:    chn,
		Exprs:    exprs,
		UserData: []byte(userTag),
	}
	c.AddRule(r)
	if err := c.Flush(); err != nil {
		log.Printf("规则 %s 添加失败: %v", userTag, err)
		return
	}
	log.Printf("规则 %s 添加成功", userTag)
}

// 删除用户规则
func deleteUserRules(userTag string) {
	tbl, chn, c := getTableAndChain()
	if c == nil || tbl == nil || chn == nil {
		log.Printf("deleteUserRules: 无法获取 table/chain")
		return
	}

	rules, err := c.GetRules(tbl, chn)
	if err != nil {
		log.Printf("GetRules 失败: %v", err)
		return
	}
	for _, r := range rules {
		if string(r.UserData) == userTag {
			c.DelRule(r)
			log.Printf("删除规则: %s", userTag)
		}
	}
	if err := c.Flush(); err != nil {
		log.Printf("deleteUserRules Flush 失败: %v", err)
	}
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
func addSimpleTcp22Rule() error {
	c := &nftables.Conn{}
	tbl := c.AddTable(&nftables.Table{
		Family: nftables.TableFamilyIPv4,
		Name:   "testtbl",
	})
	chn := c.AddChain(&nftables.Chain{
		Name:     "input",
		Table:    tbl,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookInput,
		Priority: nftables.ChainPriorityFilter,
	})

	// 匹配 TCP
	exprs := []expr.Any{
		// l4proto == tcp
		&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1},
		&expr.Cmp{Register: 1, Op: expr.CmpOpEq, Data: []byte{unix.IPPROTO_TCP}},

		// dport == 22
		&expr.Payload{
			DestRegister: 1,
			Base:         expr.PayloadBaseTransportHeader,
			Offset:       2,
			Len:          2,
		},
		&expr.Cmp{
			Register: 1,
			Op:       expr.CmpOpEq,
			Data:     []byte{0x00, 0x16}, // 22
		},

		// accept
		&expr.Verdict{Kind: expr.VerdictAccept},
	}

	c.AddRule(&nftables.Rule{
		Table: tbl,
		Chain: chn,
		Exprs: exprs,
	})

	return c.Flush()
}

func main() {

	if err := addSimpleTcp22Rule(); err != nil {
		log.Fatal(err)
	}
	log.Println("OK: added simple tcp/22 rule")

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
						deleteUserRules(userTag)
					}
				}
			}

			onlineUsers = currentUsers
			time.Sleep(refreshInterval)
		}
	}()
	select {}
}
