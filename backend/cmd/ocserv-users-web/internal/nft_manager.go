package internal

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// -------------------- 常量定义 --------------------
const (
	filterTableName        = "vpn_filter"
	filterForwardChainName = "vpn_forward"
	filterInputChainName   = "vpn_input"
)

// 根据 ip 和 tolocal 参数决定使用哪个链
func getChain(ip string, tolocal bool) (chain string) {
	if isLocalIP(ip) || tolocal { // 如果目标是本机，或者强制认为是本机
		return filterInputChainName
	}
	return filterForwardChainName
}

// -------------------- 全局状态 --------------------

var (
	onlineUsers = make(map[string][]string) // username -> []IPv4
	mu          sync.Mutex
)

// -------------------- Nftables Manager --------------------

// InitNftables 初始化nftables规则
func InitNftables(publicRules, inputChainRules, inputChainIpSetRules map[string][]Rule, srcIpSets []SrcIpSet) error {
	if err := addNatTableAndChain(); err != nil {
		return fmt.Errorf("确保NAT表和链失败: %v", err)
	}

	if err := flushFilterTableAndChain(filterTableName, filterForwardChainName, filterInputChainName); err != nil {
		return fmt.Errorf("清空过滤表和链失败: %v", err)
	}
	if err := addFilterTableAndChain(filterTableName, filterForwardChainName, filterInputChainName); err != nil {
		return fmt.Errorf("确保过滤表和链失败: %v", err)
	}

	// 添加 公共Public 规则
	for name, rules := range publicRules {
		for i, r := range rules {
			chain := getChain(r.DestIp, r.ToLocal)
			addNftRule(filterTableName, chain, "0.0.0.0/0", r.DestIp, r.Protocol, r.DestPort, r.Action, fmt.Sprintf("%s:%d", name, i))
		}
	}

	// 添加 InputChain 规则
	for name, rules := range inputChainRules {
		for i, r := range rules {
			addNftRule(filterTableName, filterInputChainName, r.SrcIp, r.DestIp, r.Protocol, r.DestPort, r.Action, fmt.Sprintf("%s:%d", name, i))
		}
	}

	// 创建 IP Set
	for _, srcIpSet := range srcIpSets {
		addIpSet(filterTableName, srcIpSet.Name, srcIpSet.Ips)
	}

	// 添加 InputChainIpSet 规则
	for name, rules := range inputChainIpSetRules {
		for i, r := range rules {
			addNftRulesIpSet(filterTableName, filterInputChainName, r.SrcIpSetName, r.DestIp, r.Protocol, r.DestPort, r.Action, fmt.Sprintf("%s:%d", name, i))
		}
	}

	// 重启后30分钟内允许SSH访问（由函数内部处理协程）
	addSshAccept30MinRuleAfterRestart(filterTableName, filterInputChainName)
	return nil
}

// UpdateNftablesRulesWithSessions 支持多设备，tag = username:ip:ruleIndex
func UpdateNftablesRulesWithSessions(usersDestRules map[string][]Rule) error {
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
			for j, r := range usersDestRules[username] {
				tag := fmt.Sprintf("user:%s:%s:%d", username, ip, j)
				chain := getChain(r.DestIp, r.ToLocal)
				addNftRule(filterTableName, chain, ip, r.DestIp, r.Protocol, r.DestPort, r.Action, tag)
			}
			clearConntrack(ip)
		}

		// 处理下线 IP
		for _, ip := range removed {
			for j := range usersDestRules[username] {
				tag := fmt.Sprintf("user:%s:%s:%d", username, ip, j)
				deleteNftRules(filterTableName, []string{filterForwardChainName, filterInputChainName}, tag)
			}
			clearConntrack(ip)
		}
	}

	// 处理完全下线用户
	for username, oldIPs := range onlineUsers {
		if _, ok := current[username]; !ok {
			log.Printf("[diff] 用户 %s 完全下线，移除所有 IP: %v", username, oldIPs)
			for _, ip := range oldIPs {
				for j := range usersDestRules[username] {
					tag := fmt.Sprintf("user:%s:%s:%d", username, ip, j)
					deleteNftRules(filterTableName, []string{filterForwardChainName, filterInputChainName}, tag)
				}
				clearConntrack(ip)
			}
		}
	}

	// 更新全局在线用户映射
	onlineUsers = current
	return nil
}

// RunNftablesManager 启动nftables管理器，定期更新规则
func RunNftablesManager(ctx context.Context, refresh time.Duration) {
	ticker := time.NewTicker(refresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[nft] 停止nftables管理器")
			return
		case <-ticker.C:
			if err := UpdateNftablesRulesWithSessions(Global_UsersRules); err != nil {
				log.Printf("[nft] 更新规则失败: %v", err)
				continue
			}
		}
	}
}
