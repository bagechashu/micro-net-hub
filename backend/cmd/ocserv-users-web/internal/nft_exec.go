package internal

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"strings"
)

func addNatTableAndChain() error {
	table := "nat"
	chain := "POSTROUTING"
	// nat POSTROUTING 链默认允许所有出站流量
	natChainAttrs := []string{"{", "type", "nat", "hook", "postrouting", "priority", "srcnat;", "policy", "accept;", "}"}
	if err := addNftTableAndChain(table, chain, natChainAttrs); err != nil {
		return err
	}

	// nat POSTROUTING 添加 masquerade 规则
	out, err := exec.Command("nft", "list", "chain", "ip", table, chain).CombinedOutput()
	if err != nil {
		return err
	}
	if !strings.Contains(string(out), "masquerade") {
		return exec.Command("nft", "add", "rule", "ip", table, chain, "masquerade").Run()
	}
	return nil
}

func addFilterTableAndChain(tableName, forwardChainName, inputChainName string) error {
	establishedComment := "allow_return_traffic"
	allowLoopbackComment := "allow_loopback"

	// 确保 filter表 和 forward, input 链存在
	// forward 链默认 drop 所有流量
	forwardChainAttrs := []string{"{", "type", "filter", "hook", "forward", "priority", "filter;", "policy", "drop;", "}"}
	if err := addNftTableAndChain(tableName, forwardChainName, forwardChainAttrs); err != nil {
		return err
	}
	// input 链默认 drop 所有流量
	inputChainAttrs := []string{"{", "type", "filter", "hook", "input", "priority", "filter;", "policy", "drop;", "}"}
	if err := addNftTableAndChain(tableName, inputChainName, inputChainAttrs); err != nil {
		return err
	}

	// forward, input 添加已建立连接放行规则
	if err := addEstablishedRule(tableName, forwardChainName, establishedComment); err != nil {
		return err
	}
	if err := addEstablishedRule(tableName, inputChainName, establishedComment); err != nil {
		return err
	}

	// vpn_input 添加允许 loopback
	if err := addLoopbackRule(tableName, inputChainName, allowLoopbackComment); err != nil {
		return err
	}

	// vpn_input 添加允许 443 端口访问规则
	// allowOcserv443Comment := "allow_ocserv_443"
	//
	// if err := exec.Command("nft", "add", "rule", "ip", tableName, filterInputChainName,
	// 	"tcp", "dport", "443", "accept", "comment", fmt.Sprintf("\"%s\"", allowOcserv443Comment)).Run(); err != nil {
	// 	return err
	// }

	return nil
}

func flushFilterTableAndChain(tableName, forwardChainName, inputChainName string) error {
	// 清空 vpn_forward 链规则
	if err := flushChain(tableName, forwardChainName); err != nil {
		return err
	}

	// 清空 vpn_input 链规则
	if err := flushChain(tableName, inputChainName); err != nil {
		return err
	}
	return nil
}

// addNftTableAndChain 添加 nftables 表和链
func addNftTableAndChain(tableName, chainName string, chainAttrs []string) error {
	if err := exec.Command("nft", "list", "table", "ip", tableName).Run(); err != nil {
		if err := exec.Command("nft", "add", "table", "ip", tableName).Run(); err != nil {
			return err
		}
	}
	if err := exec.Command("nft", "list", "chain", "ip", tableName, chainName).Run(); err != nil {
		args := []string{"add", "chain", "ip", tableName, chainName}
		args = append(args, chainAttrs...)
		if err := exec.Command("nft", args...).Run(); err != nil {
			return err
		}
	}
	return nil
}

// flushChain 清空指定链规则
func flushChain(tableName, chainName string) error {
	// 如果链存在，则 flush
	if err := exec.Command("nft", "list", "chain", "ip", tableName, chainName).Run(); err == nil {
		if err := exec.Command("nft", "flush", "chain", "ip", tableName, chainName).Run(); err != nil {
			return err
		}
	}
	// 链不存在，直接返回 nil
	return nil
}

// addEstablishedRule 添加已建立连接放行规则
func addEstablishedRule(tableName, chainName, comment string) error {
	if err := exec.Command("nft", "insert", "rule", "ip", tableName, chainName,
		"ct", "state", "established,related", "accept", "comment", fmt.Sprintf("\"%s\"", comment)).Run(); err != nil {
		return err
	}
	return nil
}

// addLoopbackRule 添加允许 loopback 访问规则
func addLoopbackRule(tableName, chainName, comment string) error {
	if err := exec.Command("nft", "add", "rule", "ip", tableName, chainName,
		"iif", "lo", "accept", "comment", fmt.Sprintf("\"%s\"", comment)).Run(); err != nil {
		return err
	}
	return nil
}

// addSshAccept30MinRuleAfterRestart 重启后30分钟内允许SSH访问（由函数内部处理协程）
func addSshAccept30MinRuleAfterRestart(tableName, inputChainName string) {
	go func() {
		tmpSshAcceptComment := "tmp_allow_ssh"
		if err := exec.Command("nft", "add", "rule", "ip", tableName, inputChainName,
			"tcp", "dport", "22", "accept", "comment", fmt.Sprintf("\"%s\"", tmpSshAcceptComment)).Run(); err != nil {
			log.Printf("[nft] 添加临时SSH放行规则失败: %v", err)
			return
		}
		time.AfterFunc(30*time.Minute, func() {
			deleteNftRules(tableName, []string{inputChainName}, tmpSshAcceptComment)
			log.Println("[nft] 已移除临时SSH放行规则")
		})
	}()
}

// addNftRule 添加自定义规则到指定链
// srcIP: 源IP地址
// dstIP: 目标IP地址
// protocol: 协议(tcp/udp/icmp)
// port: 端口号(0表示所有端口)
// toLocal: 是否要添加到 input 链(目标是本机)
// tag: 规则标签，用于后续删除
func addNftRule(table, chain, srcIP, dstIP string, protocol ProtocolType, dstport uint16, action ActionType, tag string) {
	// 默认使用 add
	args := []string{"add", "rule", "ip", table, chain}
	if action == ActionDrop {
		// 尝试找到已建立连接规则(established,related)的handle；如果找到则 add 到其后面，否则insert到链首.
		establishKeyword := "established,related"
		hs, err := getNftRuleHandle(table, chain, establishKeyword)
		if err != nil {
			log.Printf("[nft] 查找 %s 规则位置失败: %v", establishKeyword, err)
			args = []string{"insert", "rule", "ip", table, chain}
		} else if len(hs) > 0 {
			args = []string{"add", "rule", "ip", table, chain, "handle", hs[0]}
		} else {
			args = []string{"insert", "rule", "ip", table, chain}
		}
	}

	args = append(args,
		"ip", "saddr", srcIP,
		"ip", "daddr", dstIP,
	)

	switch protocol {
	case ProtocolTcp, ProtocolUdp:
		if dstport > 0 {
			args = append(args, string(protocol), "dport", fmt.Sprint(dstport))
		} else {
			args = append(args, "meta", "l4proto", string(protocol))
		}
	case ProtocolIcmp:
		args = append(args, "meta", "l4proto", string(protocol))
	default:
		log.Printf("[nft] 当前不支持协议 %s", protocol)
		return
	}

	args = append(args, string(action))
	args = append(args, "comment", fmt.Sprintf("\"%s\"", tag))

	cmd := exec.Command("nft", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[nft] 添加规则失败 %s: %v (%s)", tag, err, out)
	} else {
		log.Printf("[nft] 执行命令: %s", strings.Join(cmd.Args, " "))
	}
}

func deleteNftRules(table string, chains []string, tag string) {
	// 遍历 forward/input 两个链
	for _, chain := range chains {
		if hs, err := getNftRuleHandle(table, chain, tag); err != nil || len(hs) != 0 {
			for _, h := range hs {
				delCmd := exec.Command("nft", "delete", "rule", "ip", table, chain, "handle", h)
				if outDel, err := delCmd.CombinedOutput(); err != nil {
					log.Printf("[nft] 删除规则失败: %v (%s)", err, outDel)
				} else {
					log.Printf("[nft] 执行命令: %s", strings.Join(delCmd.Args, " "))
				}
			}
		}
	}
}

// AddIpSetRules 添加允许指定 IP 集合访问 ocserv 443 的规则
func addNftRulesIpSet(table, chain, srcIpSetname, dstIP string, protocol ProtocolType, dstport uint16, action ActionType, tag string) {
	// 默认使用 add
	args := []string{"add", "rule", "ip", table, chain}
	if action == ActionDrop {
		// 尝试找到已建立连接规则(established,related)的handle；如果找到则 add 到其后面，否则insert到链首.
		establishKeyword := "established,related"
		hs, err := getNftRuleHandle(table, chain, establishKeyword)
		if err != nil {
			log.Printf("[nft] 查找 %s 规则位置失败: %v", establishKeyword, err)
			args = []string{"insert", "rule", "ip", table, chain}
		} else if len(hs) > 0 {
			args = []string{"add", "rule", "ip", table, chain, "handle", hs[0]}
		} else {
			args = []string{"insert", "rule", "ip", table, chain}
		}
	}

	args = append(args,
		"ip", "saddr", fmt.Sprintf("@%s", srcIpSetname),
		"ip", "daddr", dstIP,
	)

	switch protocol {
	case ProtocolTcp, ProtocolUdp:
		if dstport > 0 {
			args = append(args, string(protocol), "dport", fmt.Sprint(dstport))
		} else {
			args = append(args, "meta", "l4proto", string(protocol))
		}
	case ProtocolIcmp:
		args = append(args, "meta", "l4proto", string(protocol))
	default:
		log.Printf("[nft] 当前不支持协议 %s", protocol)
		return
	}

	args = append(args, string(action))
	args = append(args, "comment", fmt.Sprintf("\"%s\"", tag))

	cmd := exec.Command("nft", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[nft] 添加规则失败: %v (%s)", err, out)
	} else {
		log.Printf("[nft] 执行命令: %s", strings.Join(cmd.Args, " "))
	}

}

// DeleteIpSetRules 删除指定 set 对应的规则
func deleteNftRulesIpSet(table string, chains []string, srcIpSets []string) {
	// 遍历 forward/input 两个链
	for _, chain := range chains {
		for _, setName := range srcIpSets {
			if hs, err := getNftRuleHandle(table, chain, fmt.Sprintf("@%s", setName)); err != nil || len(hs) != 0 {
				for _, h := range hs {
					delCmd := exec.Command("nft", "delete", "rule", "ip", table, chain, "handle", h)
					if outDel, err := delCmd.CombinedOutput(); err != nil {
						log.Printf("[nft] 删除规则失败: %v (%s)", err, outDel)
					} else {
						log.Printf("[nft] 执行命令: %s", strings.Join(delCmd.Args, " "))
					}
				}
			}
		}
	}
}

// addIpSet 在 nftables 中创建 IP 集合
func addIpSet(table, setName string, ips []string) {
	// 判断 set 是否存在
	deleteIpSet(table, setName)

	// 创建 set
	cmdAdd := exec.Command("nft", "add", "set", "ip", table, setName,
		"{", "type", "ipv4_addr;", "flags", "interval;", "}")
	if out, err := cmdAdd.CombinedOutput(); err != nil {
		log.Printf("[nft] 创建 set %s 失败: %v (%s)", setName, err, out)
	}

	// 批量添加元素
	ipsLen := len(ips)
	if ipsLen > 0 {
		args := []string{"add", "element", "ip", table, setName, "{"}
		for i, ip := range ips {
			args = append(args, ip)
			if i != ipsLen-1 {
				args = append(args, ",")
			}
		}
		args = append(args, "}")
		cmdEl := exec.Command("nft", args...)
		if out, err := cmdEl.CombinedOutput(); err != nil {
			log.Printf("[nft] 添加元素到 set %s 失败: %v (%s)", setName, err, out)
		}
	}

	log.Printf("[nft] 创建 set %s 完成，包含 %d 个 IP", setName, ipsLen)
}

// deleteIpSet 在 nftables 中删除 IP 集合
func deleteIpSet(table, setName string) {
	// 判断 set 是否存在
	check := exec.Command("nft", "list", "set", "ip", table, setName)
	if err := check.Run(); err == nil {
		// set 已存在，先 flush
		cmdFlush := exec.Command("nft", "flush", "set", "ip", table, setName)
		if out, err := cmdFlush.CombinedOutput(); err != nil {
			log.Printf("[nft] flush set %s 失败: %v (%s)", setName, err, out)
		}
	}
}

func getNftRuleHandle(table, chain string, keyword string) (handles []string, err error) {
	cmd := exec.Command("nft", "-a", "list", "chain", "ip", table, chain)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("[nft] 列出规则失败(%s): %v", chain, err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, keyword) && strings.Contains(line, "# handle") {
			h := strings.TrimSpace(line[strings.LastIndex(line, "# handle ")+9:])
			handles = append(handles, h)
		}
	}

	return handles, nil
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
