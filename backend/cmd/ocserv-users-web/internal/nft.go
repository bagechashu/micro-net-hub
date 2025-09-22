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
	establishedComment     = "allow return traffic"
)

// -------------------- 全局状态 --------------------

var (
	onlineUsers = make(map[string][]string) // username -> []IPv4
	mu          sync.Mutex
)

// -------------------- Nftables Manager --------------------

func InitNftables(publicRules []RuleConfig) error {
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
		addNftRule(fmt.Sprintf("public:%d", i), "0.0.0.0/0", r)
	}
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
				addNftRule(tag, ip, r)
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

func addNftRule(tag, srcIP string, r RuleConfig) {
	proto := strings.ToLower(r.Protocol)

	// 自动识别目标是否是本机
	chain := filterForwardChainName
	if isLocalIP(r.IP) || r.ToLocal { // 如果目标是本机，或者强制认为是本机
		chain = filterInputChainName
	}

	args := []string{
		"add", "rule", "ip", filterTableName, chain,
		"ip", "saddr", srcIP, "ip", "daddr", r.IP,
	}

	switch proto {
	case "tcp", "udp":
		if r.Port > 0 {
			args = append(args, proto, "dport", fmt.Sprint(r.Port))
		} else {
			args = append(args, "meta", "l4proto", proto)
		}
	case "icmp":
		args = append(args, "meta", "l4proto", proto)
	default:
		log.Printf("[nft] 当前不支持协议 %s", r.Protocol)
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
