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
	filterTableName    = "vpn_filter"
	filterChainName    = "vpn_forward"
	establishedComment = "allow return traffic"
)

// -------------------- 全局状态 --------------------

var (
	onlineUsers  = make(map[string][]string) // username -> []IPv4
	mu           sync.Mutex
	userSessions = make(map[string]time.Time) // username -> login time
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
	if err := exec.Command("nft", "list", "table", "ip", filterTableName).Run(); err == nil {
		if err := exec.Command("nft", "flush", "chain", "ip", filterTableName, filterChainName).Run(); err != nil {
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
	if err := exec.Command("nft", "list", "chain", "ip", filterTableName, filterChainName).Run(); err != nil {
		if err := exec.Command("nft", "add", "chain", "ip", filterTableName, filterChainName,
			"{", "type", "filter", "hook", "forward", "priority", "filter;", "policy", "drop;", "}").Run(); err != nil {
			return err
		}
	}

	// 添加已建立连接放行规则
	if err := exec.Command("nft", "add", "rule", "ip", filterTableName, filterChainName,
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

// UpdateNftablesRulesWithSessions
// 支持同一用户多个设备
func UpdateNftablesRulesWithSessions(userRules map[string][]RuleConfig) error {
	sessions, err := GetSessions()
	if err != nil {
		log.Printf("[nft] 获取会话失败: %v", err)
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	current := make(map[string][]string) // username -> []IPv4
	currentSessions := make(map[string]time.Time)

	// 构建当前在线用户映射
	for _, s := range sessions {
		current[s.Username] = append(current[s.Username], s.IPv4)
		currentSessions[s.Username] = time.Now()
	}

	// 处理每个在线用户
	for username, ips := range current {
		oldIPs, exists := onlineUsers[username]

		if !exists {
			// 新用户登录（任何设备）
			userSessions[username] = time.Now()
			log.Printf("[user] 用户 %s 首次登录，IP列表: %v，时间: %s", username, ips, userSessions[username].Format("2006-01-02 15:04:05"))
			for i, ip := range ips {
				if ip != "" {
					for j, r := range userRules[username] {
						addNftRule(fmt.Sprintf("user:%s:%d:%d", username, i, j), ip, r)
					}
				}
			}
		} else {
			// 检查是否有IP变更
			needUpdate := false

			// 检查是否有新IP（新设备登录）
			for _, newIP := range ips {
				found := false
				for _, oldIP := range oldIPs {
					if newIP == oldIP {
						found = true
						break
					}
				}
				if !found && newIP != "" {
					needUpdate = true
					log.Printf("[user] 用户 %s 新设备登录，新增IP: %s，时间: %s", username, newIP, time.Now().Format("2006-01-02 15:04:05"))
				}
			}

			// 检查是否有IP变更（设备重新连接）
			for _, oldIP := range oldIPs {
				found := false
				for _, newIP := range ips {
					if oldIP == newIP {
						found = true
						break
					}
				}
				if !found {
					needUpdate = true
					log.Printf("[user] 用户 %s 设备断开，移除IP: %s，时间: %s", username, oldIP, time.Now().Format("2006-01-02 15:04:05"))
				}
			}

			if needUpdate {
				// 删除该用户所有规则
				for i := range oldIPs {
					for j := range userRules[username] {
						deleteNftRules(fmt.Sprintf("user:%s:%d:%d", username, i, j))
					}
				}

				// 为所有当前IP重新添加规则
				for i, ip := range ips {
					if ip != "" {
						for j, r := range userRules[username] {
							addNftRule(fmt.Sprintf("user:%s:%d:%d", username, i, j), ip, r)
						}
					}
				}
			}
		}
	}

	// 处理完全离线的用户
	for username, oldIPs := range onlineUsers {
		if _, ok := current[username]; !ok {
			// 用户完全离线
			if loginTime, exists := userSessions[username]; exists {
				log.Printf("[user] 用户 %s 完全离线，登录时间: %s，离线时间: %s，时长: %v",
					username,
					loginTime.Format("2006-01-02 15:04:05"),
					time.Now().Format("2006-01-02 15:04:05"),
					time.Since(loginTime))
				delete(userSessions, username)
			}
			// 删除该用户所有规则
			for i := range oldIPs {
				for j := range userRules[username] {
					deleteNftRules(fmt.Sprintf("user:%s:%d:%d", username, i, j))
				}
			}
		}
	}

	onlineUsers = current
	return nil
}

func addNftRule(tag, srcIP string, r RuleConfig) {
	proto := strings.ToLower(r.Protocol)
	args := []string{
		"add", "rule", "ip", filterTableName, filterChainName,
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
		args = append(args, proto)
	default:
		log.Printf("[nft] 忽略未知协议 %s", r.Protocol)
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
	cmd := exec.Command("nft", "-a", "list", "chain", "ip", filterTableName, filterChainName)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("[nft] 列出规则失败: %v", err)
		return
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, tag) && strings.Contains(line, "# handle") {
			handle := strings.TrimSpace(line[strings.LastIndex(line, "# handle ")+9:])
			delCmd := exec.Command("nft", "delete", "rule", "ip", filterTableName, filterChainName, "handle", handle)
			if delOut, err := delCmd.CombinedOutput(); err != nil {
				log.Printf("[nft] 删除规则失败 %s: %v (%s)", tag, err, delOut)
			} else {
				log.Printf("[nft] 执行命令: %s", strings.Join(delCmd.Args, " "))
			}
		}
	}
}
