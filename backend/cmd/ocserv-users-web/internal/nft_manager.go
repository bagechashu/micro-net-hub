package internal

import (
	"context"
	"fmt"
	"log"
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

// UpdateUsersNftablesRules 支持多设备，tag = username:ip:ruleIndex
func UpdateUsersNftablesRules(usersDestRules map[string][]Rule) error {
	// 更新在线用户状态和 nftables 规则
	addedSessions, removedSessions, err := checkOcSessions()
	if err != nil {
		return err
	}
	updateUsersNftRules(addedSessions, removedSessions, usersDestRules)
	return nil
}

// UpdateUsersNftablesRules 支持多设备，tag = username:ip:ruleIndex
func InitUsersNftablesRules(usersDestRules map[string][]Rule) error {
	offlineAllSessions()
	// 更新在线用户状态和 nftables 规则
	addedSessions, _, err := checkOcSessions()
	if err != nil {
		return err
	}
	updateUsersNftRules(addedSessions, nil, usersDestRules)
	return nil
}

// updateOnlineUsers 根据当前在线用户映射更新全局状态和 nftables 规则
func updateUsersNftRules(addedSessions, removedSessions map[string][]string, usersDestRules map[string][]Rule) {
	// 先下线再上线，避免冲突
	for username, removed := range removedSessions {
		if len(removed) > 0 {
			handleRemovedIPs(username, removed, usersDestRules)
		}
	}

	for username, added := range addedSessions {
		if len(added) > 0 {
			handleAddedIPs(username, added, usersDestRules)
		}
	}
}

// handleAddedIPs 处理新增 IP：添加 nftables 规则并清除连接跟踪
func handleAddedIPs(username string, added []string, usersDestRules map[string][]Rule) {
	for _, ip := range added {
		for j, r := range usersDestRules[username] {
			tag := fmt.Sprintf("user:%s:%s:%d", username, ip, j)
			chain := getChain(r.DestIp, r.ToLocal)
			addNftRule(filterTableName, chain, ip, r.DestIp, r.Protocol, r.DestPort, r.Action, tag)
		}
		clearConntrack(ip)
	}
}

// handleRemovedIPs 处理下线 IP：删除 nftables 规则并清除连接跟踪
func handleRemovedIPs(username string, removed []string, usersDestRules map[string][]Rule) {
	for _, ip := range removed {
		for j := range usersDestRules[username] {
			tag := fmt.Sprintf("user:%s:%s:%d", username, ip, j)
			deleteNftRules(filterTableName, []string{filterForwardChainName, filterInputChainName}, tag)
		}
		clearConntrack(ip)
	}
}

// RunNftablesManager 启动nftables管理器，定期更新规则
func RunNftablesManager(ctx context.Context, refresh time.Duration) {
	log.Printf("[nft] 启动 nftables 管理器 (refresh=%s)", refresh.String())
	ticker := time.NewTicker(refresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[nft] 停止nftables管理器")
			return
		case <-ticker.C:
			if err := UpdateUsersNftablesRules(globalUsersRules); err != nil {
				log.Printf("[nft] 更新规则失败: %v", err)
				continue
			}
		}
	}
}
