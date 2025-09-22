package internal

import (
	"fmt"
	"log"
	"os/exec"
)

// CheckDependencies 检查系统依赖命令是否存在
func CheckDependencies() error {
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
