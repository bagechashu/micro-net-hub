package internal

import (
	"fmt"
	"log"
	"os/exec"

	"strings"
)

// CreateIpSet 在 nftables 中创建国家 IP 集合
func CreateIpSet(table, setName string, ips []string) error {
	// 判断 set 是否存在
	check := exec.Command("nft", "list", "set", "ip", table, setName)
	if err := check.Run(); err == nil {
		// set 已存在，先 flush
		cmdFlush := exec.Command("nft", "flush", "set", "ip", table, setName)
		if out, err := cmdFlush.CombinedOutput(); err != nil {
			return fmt.Errorf("flush set %s 失败: %v (%s)", setName, err, out)
		}
	}

	// 创建 set
	cmdAdd := exec.Command("nft", "add", "set", "ip", table, setName,
		"{", "type", "ipv4_addr;", "flags", "interval;", "}")
	if out, err := cmdAdd.CombinedOutput(); err != nil {
		return fmt.Errorf("创建 set %s 失败: %v (%s)", setName, err, out)
	}

	// 批量添加元素
	if len(ips) > 0 {
		args := []string{"add", "element", "ip", table, setName, "{"}
		for i, ip := range ips {
			args = append(args, ip)
			if i != len(ips)-1 {
				args = append(args, ",")
			}
		}
		args = append(args, "}")
		cmdEl := exec.Command("nft", args...)
		if out, err := cmdEl.CombinedOutput(); err != nil {
			return fmt.Errorf("添加元素到 set %s 失败: %v (%s)", setName, err, out)
		}
	}

	log.Printf("[nft] 创建 set %s 完成，包含 %d 个 IP", setName, len(ips))
	return nil
}

// AddIpSetRules 添加允许指定 IP 集合访问 ocserv 443 的规则
func AddIpSetRules(table, chain string, ipSets []string) error {
	for _, setName := range ipSets {
		args := []string{
			"add", "rule", "ip", table, chain,
			"tcp", "dport", "443",
			"ip", "saddr", fmt.Sprintf("@%s", setName),
			"accept",
		}
		cmd := exec.Command("nft", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("添加规则失败: %v (%s)", err, out)
		}
		log.Printf("[nft] 添加规则: 允许 %s 访问 443", setName)
	}
	return nil
}

// DeleteIpSetRules 删除指定 set 对应的规则
func DeleteIpSetRules(table, chain string, ipSets []string) {
	for _, setName := range ipSets {
		cmdList := exec.Command("nft", "-a", "list", "chain", "ip", table, chain)
		out, err := cmdList.Output()
		if err != nil {
			log.Printf("[nft] list chain 错误: %v", err)
			continue
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, fmt.Sprintf("@%s", setName)) && strings.Contains(line, "# handle") {
				handle := strings.TrimSpace(line[strings.LastIndex(line, "# handle ")+9:])
				cmdDel := exec.Command("nft", "delete", "rule", "ip", table, chain, "handle", handle)
				if outDel, err := cmdDel.CombinedOutput(); err != nil {
					log.Printf("[nft] 删除规则失败: %v (%s)", err, outDel)
				} else {
					log.Printf("[nft] 删除规则 handle %s 成功", handle)
				}
			}
		}
	}
}
