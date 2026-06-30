package internal

// DebugMode 控制是否启用调试模式
// 仅在调试模式下打印详细的执行命令日志
var DebugMode bool = false

// SetDebugMode 设置全局 debug 模式
func SetDebugMode(debug bool) {
	DebugMode = debug
}

// IsDebugMode 返回当前是否启用调试模式
func IsDebugMode() bool {
	return DebugMode
}
