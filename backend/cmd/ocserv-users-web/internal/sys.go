package internal

import (
	"fmt"
	"log"
	"net"
	"os/exec"
)

// localAddrs 本机地址缓存
var localAddrs []net.IP

func init() {
	var err error
	localAddrs, err = getLocalIPv4s()
	if err != nil {
		log.Printf("[sys] 获取本机IP失败: %v", err)
	} else {
		log.Printf("[sys] 本机IPv4地址: %v", localAddrs)
	}

	// 检查系统依赖
	if err := checkDependencies(); err != nil {
		log.Fatalf("[sys] 检查系统依赖失败: %v", err)
	}
}

func getLocalIPv4s() ([]net.IP, error) {
	var ips []net.IP
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces: %w", err)
	}
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					ips = append(ips, ip4)
				}
			}
		}
	}
	return ips, nil
}

func isLocalIP(ip string) bool {
	// 先尝试按 CIDR 解析，比如 10.12.0.216/32
	if parsedIP, _, err := net.ParseCIDR(ip); err == nil {
		ip = parsedIP.String()
	}

	dst := net.ParseIP(ip)
	if dst == nil {
		log.Printf("[sys] 规则中的 IP 地址无效: %s", ip)
		return false
	}

	for _, local := range localAddrs {
		if local.Equal(dst) {
			return true
		}
	}
	return false
}

// checkDependencies 检查系统依赖命令是否存在
func checkDependencies() error {
	commands := []string{"nft", "conntrack", "occtl"}
	missing := []string{}

	for _, cmd := range commands {
		if _, err := exec.LookPath(cmd); err != nil {
			missing = append(missing, cmd)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("缺少必要命令: %v", missing)
	}

	log.Println("[init] 系统依赖检查通过: nft, conntrack, occtl 已安装")
	return nil
}
