package internal

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
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
		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("[sys] 获取接口地址失败 (%s): %v", iface.Name, err)
			continue
		}
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

// clearConntrack 清理指定源IP的连接跟踪条目
func clearConntrack(ip string) {
	cmd := exec.Command("conntrack", "-D", "-s", ip)
	out, err := cmd.CombinedOutput()
	output := string(out)

	if err != nil {
		// 特殊情况：没有条目被删除
		if strings.Contains(output, "0 flow entries have been deleted") {
			log.Printf("[sys] 执行命令: %s (没有匹配的条目)", strings.Join(cmd.Args, " "))
			return
		}
		// 其他错误才是真的失败
		log.Printf("[sys] 执行命令: %s 清理失败: %v (%s)", strings.Join(cmd.Args, " "), err, output)
		return
	}

	log.Printf("[sys] 执行命令: %s", strings.Join(cmd.Args, " "))
}
