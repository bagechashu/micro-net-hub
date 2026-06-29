package internal

import (
	"fmt"
	"runtime"
)

// 这些变量将在编译时通过 -ldflags 注入
var (
	Version   = "dev"       // 版本号，如 v1.0.0
	GitCommit = "none"      // Git Commit Hash
	BuildTime = "unknown"   // 编译时间
	GoVersion = runtime.Version() // Go 版本
)

// GetVersionInfo 返回格式化的版本信息字符串
func GetVersionInfo() string {
	return fmt.Sprintf(
		"Version: %s\nGit Commit: %s\nBuild Time: %s\nGo Version: %s",
		Version,
		GitCommit,
		BuildTime,
		GoVersion,
	)
}