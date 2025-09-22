package internal

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// -------------------- 常量定义 --------------------
const (
	filterTableName        = "vpn_filter"
	filterForwardChainName = "vpn_forward"
	filterInputChainName   = "vpn_input"
)

// -------------------- 全局状态 --------------------

var (
	onlineUsers = make(map[string][]string) // username -> []IPv4
	mu          sync.Mutex
)

// -------------------- Nftables Manager --------------------

// InitNftables 初始化nftables规则
func InitNftables(publicRules []RuleConfig, inputChainRules map[string][]RuleConfig) error {
	if err := ensureNatTableAndChain(); err != nil {
		return fmt.Errorf("确保NAT表和链失败: %v", err)
	}

	if err := flushFilterTableAndChain(); err != nil {
		return fmt.Errorf("清空过滤表和链失败: %v", err)
	}
	if err := ensureFilterTableAndChain(); err != nil {
		return fmt.Errorf("确保过滤表和链失败: %v", err)
	}

	// 添加公共规则
	for i, r := range publicRules {
		addNftRule("0.0.0.0/0", r.IP, r.Protocol, r.Port, r.ToLocal, fmt.Sprintf("public:%d", i))
	}

	// 添加 InputChain 规则
	i := 0
	for srcIP, rules := range inputChainRules {
		for _, r := range rules {
			addNftRule(srcIP, r.IP, r.Protocol, r.Port, r.ToLocal, fmt.Sprintf("input:%d", i))
			i++
		}
	}

	go enableSshAccept30MinAfterRestart()
	return nil
}

func ensureNatTableAndChain() error {
	if err := exec.Command("nft", "list", "table", "ip", "nat").Run(); err != nil {
		if err := exec.Command("nft", "add", "table", "ip", "nat").Run(); err != nil {
			return err
		}
	}
	if err := exec.Command("nft", "list", "chain", "ip", "nat", "POSTROUTING").Run(); err != nil {
		if err := exec.Command("nft", "add", "chain", "ip", "nat", "POSTROUTING",
			"{", "type", "nat", "hook", "postrouting", "priority", "srcnat;", "policy", "accept;", "}").Run(); err != nil {
			return err
		}
	}
	out, err := exec.Command("nft", "list", "chain", "ip", "nat", "POSTROUTING").Output()
	if err != nil {
		return err
	}
	if !strings.Contains(string(out), "masquerade") {
		return exec.Command("nft", "add", "rule", "ip", "nat", "POSTROUTING", "masquerade").Run()
	}
	return nil
}

func flushFilterTableAndChain() error {
	// 清空 vpn_forward 链规则
	if err := exec.Command("nft", "list", "chain", "ip", filterTableName, filterForwardChainName).Run(); err == nil {
		if err := exec.Command("nft", "flush", "chain", "ip", filterTableName, filterForwardChainName).Run(); err != nil {
			return err
		}
	}

	// 清空 vpn_input 链规则
	if err := exec.Command("nft", "list", "chain", "ip", filterTableName, filterInputChainName).Run(); err == nil {
		if err := exec.Command("nft", "flush", "chain", "ip", filterTableName, filterInputChainName).Run(); err != nil {
			return err
		}
	}
	return nil
}

func ensureFilterTableAndChain() error {
	establishedComment := "allow_return_traffic"
	allowLoopbackComment := "allow_loopback"
	allowOcserv443Comment := "allow_ocserv_443"
	// 确保 vpn_filter 表存在
	if err := exec.Command("nft", "list", "table", "ip", filterTableName).Run(); err != nil {
		if err := exec.Command("nft", "add", "table", "ip", filterTableName).Run(); err != nil {
			return err
		}
	}

	// 确保 vpn_forward 链存在
	if err := exec.Command("nft", "list", "chain", "ip", filterTableName, filterForwardChainName).Run(); err != nil {
		if err := exec.Command("nft", "add", "chain", "ip", filterTableName, filterForwardChainName,
			"{", "type", "filter", "hook", "forward", "priority", "filter;", "policy", "drop;", "}").Run(); err != nil {
			return err
		}
	}

	// vpn_forward 添加已建立连接放行规则
	if err := exec.Command("nft", "add", "rule", "ip", filterTableName, filterForwardChainName,
		"ct", "state", "established,related", "accept", "comment", fmt.Sprintf("\"%s\"", establishedComment)).Run(); err != nil {
		return err
	}

	// 确保 vpn_input 链存在
	if err := exec.Command("nft", "list", "chain", "ip", filterTableName, filterInputChainName).Run(); err != nil {
		if err := exec.Command("nft", "add", "chain", "ip", filterTableName, filterInputChainName,
			"{", "type", "filter", "hook", "input", "priority", "filter;", "policy", "drop;", "}").Run(); err != nil {
			return err
		}
	}

	// vpn_input 添加已建立连接放行规则
	if err := exec.Command("nft", "add", "rule", "ip", filterTableName, filterInputChainName,
		"ct", "state", "established,related", "accept", "comment", fmt.Sprintf("\"%s\"", establishedComment)).Run(); err != nil {
		return err
	}

	// vpn_input 添加允许 loopback
	if err := exec.Command("nft", "add", "rule", "ip", filterTableName, filterInputChainName,
		"iif", "lo", "accept", "comment", fmt.Sprintf("\"%s\"", allowLoopbackComment)).Run(); err != nil {
		return err
	}

	// vpn_input 添加允许 443 端口访问规则
	if err := exec.Command("nft", "add", "rule", "ip", filterTableName, filterInputChainName,
		"tcp", "dport", "443", "accept", "comment", fmt.Sprintf("\"%s\"", allowOcserv443Comment)).Run(); err != nil {
		return err
	}
	return nil
}

// enableSshAccept30MinAfterRestart 重启后30分钟内允许SSH访问
func enableSshAccept30MinAfterRestart() error {
	tmpSshAcceptComment := "tmp_allow_ssh"
	if err := exec.Command("nft", "add", "rule", "ip", filterTableName, filterInputChainName,
		"tcp", "dport", "22", "accept", "comment", fmt.Sprintf("\"%s\"", tmpSshAcceptComment)).Run(); err != nil {
		return err
	}
	time.AfterFunc(30*time.Minute, func() {
		deleteNftRules(tmpSshAcceptComment)
		log.Println("[nft] 已移除临时SSH放行规则")
	})
	return nil
}

func RunNftablesManager(ctx context.Context, refresh time.Duration) {
	ticker := time.NewTicker(refresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[nft] 停止nftables管理器")
			return
		case <-ticker.C:
			if err := UpdateNftablesRulesWithSessions(GlobalUserRules); err != nil {
				log.Printf("[nft] 更新规则失败: %v", err)
				continue
			}
		}
	}
}

// UpdateNftablesRulesWithSessions 支持多设备，tag = username:ip:ruleIndex
func UpdateNftablesRulesWithSessions(userRules map[string][]RuleConfig) error {
	sessions, err := GetSessions()
	if err != nil {
		log.Printf("[nft] 获取会话失败: %v", err)
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	// 当前在线用户映射 username -> []IPv4
	current := make(map[string][]string)
	for _, s := range sessions {
		if s.Username == "" || s.IPv4 == "" {
			continue
		}
		current[s.Username] = append(current[s.Username], s.IPv4)
	}

	for username, ips := range current {
		oldIPs := onlineUsers[username]

		added, removed := diffIPs(oldIPs, ips)
		if len(added) > 0 {
			log.Printf("[diff] 用户 %s 新上线 IP: %v", username, added)
		}
		if len(removed) > 0 {
			log.Printf("[diff] 用户 %s 下线 IP: %v", username, removed)
		}

		// 处理新增 IP
		for _, ip := range added {
			for j, r := range userRules[username] {
				tag := fmt.Sprintf("user:%s:%s:%d", username, ip, j)
				addNftRule(ip, r.IP, r.Protocol, r.Port, r.ToLocal, tag)
			}
			clearConntrack(ip)
		}

		// 处理下线 IP
		for _, ip := range removed {
			for j := range userRules[username] {
				tag := fmt.Sprintf("user:%s:%s:%d", username, ip, j)
				deleteNftRules(tag)
			}
			clearConntrack(ip)
		}
	}

	// 处理完全下线用户
	for username, oldIPs := range onlineUsers {
		if _, ok := current[username]; !ok {
			log.Printf("[diff] 用户 %s 完全下线，移除所有 IP: %v", username, oldIPs)
			for _, ip := range oldIPs {
				for j := range userRules[username] {
					tag := fmt.Sprintf("user:%s:%s:%d", username, ip, j)
					deleteNftRules(tag)
				}
				clearConntrack(ip)
			}
		}
	}

	// 更新全局在线用户映射
	onlineUsers = current
	return nil
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

// addNftRule 添加自定义规则到指定链
// srcIP: 源IP地址
// dstIP: 目标IP地址
// protocol: 协议(tcp/udp/icmp)
// port: 端口号(0表示所有端口)
// toLocal: 是否要添加到 input 链(目标是本机)
// tag: 规则标签，用于后续删除
func addNftRule(srcIP, dstIP, protocol string, port uint16, toLocal bool, tag string) {
	// 自动识别目标是否是本机
	chain := filterForwardChainName
	if isLocalIP(dstIP) || toLocal { // 如果目标是本机，或者强制认为是本机
		chain = filterInputChainName
	}

	proto := strings.ToLower(protocol)
	args := []string{
		"add", "rule", "ip", filterTableName, chain,
		"ip", "saddr", srcIP, "ip", "daddr", dstIP,
	}

	switch proto {
	case "tcp", "udp":
		if port > 0 {
			args = append(args, proto, "dport", fmt.Sprint(port))
		} else {
			args = append(args, "meta", "l4proto", proto)
		}
	case "icmp":
		args = append(args, "meta", "l4proto", proto)
	default:
		log.Printf("[nft] 当前不支持协议 %s", protocol)
		return
	}

	args = append(args, "accept", "comment", fmt.Sprintf("\"%s\"", tag))

	cmd := exec.Command("nft", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[nft] 添加规则失败 %s: %v (%s)", tag, err, out)
	} else {
		log.Printf("[nft] 执行命令: %s", strings.Join(cmd.Args, " "))
	}
}

func deleteNftRules(tag string) {
	// 遍历 forward/input 两个链
	for _, chain := range []string{filterForwardChainName, filterInputChainName} {
		cmd := exec.Command("nft", "-a", "list", "chain", "ip", filterTableName, chain)
		out, err := cmd.Output()
		if err != nil {
			log.Printf("[nft] 列出规则失败(%s): %v", chain, err)
			continue
		}

		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, tag) && strings.Contains(line, "# handle") {
				handle := strings.TrimSpace(line[strings.LastIndex(line, "# handle ")+9:])
				delCmd := exec.Command("nft", "delete", "rule", "ip", filterTableName, chain, "handle", handle)
				if delOut, err := delCmd.CombinedOutput(); err != nil {
					log.Printf("[nft] 删除规则失败 %s (%s): %v (%s)", tag, chain, err, delOut)
				} else {
					log.Printf("[nft] 执行命令: %s", strings.Join(delCmd.Args, " "))
				}
			}
		}
	}
}

func clearConntrack(ip string) {
	cmd := exec.Command("conntrack", "-D", "-s", ip)
	out, err := cmd.CombinedOutput()
	output := string(out)

	if err != nil {
		// 特殊情况：没有条目被删除
		if strings.Contains(output, "0 flow entries have been deleted") {
			log.Printf("[nft] 执行命令: %s (没有匹配的条目)", strings.Join(cmd.Args, " "))
			return
		}
		// 其他错误才是真的失败
		log.Printf("[nft] 执行命令: %s 清理失败: %v (%s)", strings.Join(cmd.Args, " "), err, output)
		return
	}

	log.Printf("[nft] 执行命令: %s", strings.Join(cmd.Args, " "))
}
