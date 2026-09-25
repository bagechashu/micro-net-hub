// Package scripthook 提供基于配置的命名脚本异步调用能力。
//
// 业务模块通过 Run(ctx, name) 按名称触发脚本; 脚本以 /bin/sh 异步执行。
// 通过 WithEnv(ctx, k, v) 可将键值对注入 context, 脚本内以环境变量
// SCRIPTHOOK_<KEY> 读取。
package scripthook

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"
)

type ctxKey string

const envPrefix = "SCRIPTHOOK_"

// envKey 用于在 context 中存储脚本环境变量键值对。
const envCtxKey ctxKey = "scripthook_env"

// envPair 环境变量键值对。
type envPair struct {
	key   string
	value string
}

// Run 异步执行配置中对应名称的脚本, 立即返回。
//
// ctx 双重职责:
//  1. 超时控制: 优先取 ctx deadline, 无 deadline 则用配置 timeout-seconds(默认30s)
//  2. 数据传递: 通过 WithEnv(ctx, k, v) 注入的键值对, 以 SCRIPTHOOK_<KEY> 环境变量传入脚本
//
// 配置缺失或脚本名未找到时仅记录 warn 日志后返回, 不启动 goroutine。
// 脚本执行失败也仅记录日志, 不阻塞调用方。
func Run(ctx context.Context, name string) {
	cfg := config.Conf.ScriptHook
	if cfg == nil || cfg.Dir == "" {
		global.Log.Warnf("scripthook: 脚本调用配置缺失(skip): name=%s", name)
		return
	}

	scriptPath, ok := cfg.Scripts[name]
	if !ok || scriptPath == "" {
		global.Log.Warnf("scripthook: 脚本 %q 未在 script-hook.scripts 中配置(skip)", name)
		return
	}

	fullPath := filepath.Join(cfg.Dir, scriptPath)

	// 收集环境变量
	var envPairs []envPair
	if envs, ok := ctx.Value(envCtxKey).([]envPair); ok {
		envPairs = envs
	}

	// 计算超时
	timeout := 30 * time.Second
	if cfg.TimeoutSeconds > 0 {
		timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}
	if deadline, ok := ctx.Deadline(); ok {
		if d := time.Until(deadline); d > 0 && d < timeout {
			timeout = d
		}
	}

	go runScript(fullPath, name, timeout, envPairs)
}

// WithEnv 将键值对注入 ctx, 脚本执行时作为环境变量传入。
//
// key 自动转为大写并添加 SCRIPTHOOK_ 前缀; 脚本内可通过
// $SCRIPTHOOK_MESSAGE 等形式读取。
func WithEnv(ctx context.Context, key, value string) context.Context {
	var envs []envPair
	if existing, ok := ctx.Value(envCtxKey).([]envPair); ok {
		envs = make([]envPair, 0, len(existing)+1)
		envs = append(envs, existing...)
	}
	envs = append(envs, envPair{key: key, value: value})
	return context.WithValue(ctx, envCtxKey, envs)
}

// runScript 在 goroutine 中执行脚本, 捕获并记录输出。
func runScript(scriptPath, name string, timeout time.Duration, envPairs []envPair) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", scriptPath)

	// 注入环境变量: SCRIPTHOOK_<KEY>=<value>
	for _, p := range envPairs {
		cmd.Env = append(cmd.Env, envPrefix+toUpper(p.key)+"="+p.value)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	elapsed := time.Since(start)

	if stdoutBuf.Len() > 0 {
		global.Log.Infow("scripthook stdout",
			"script", name, "path", scriptPath, "elapsed", elapsed,
			"output", stdoutBuf.String(),
		)
	}

	if stderrBuf.Len() > 0 {
		global.Log.Warnw("scripthook stderr",
			"script", name, "path", scriptPath, "elapsed", elapsed,
			"output", stderrBuf.String(),
		)
	}

	if err != nil {
		global.Log.Errorw("scripthook: 脚本执行失败",
			"script", name, "path", scriptPath, "elapsed", elapsed, "error", err,
		)
		return
	}

	global.Log.Infow("scripthook: 脚本执行完成",
		"script", name, "path", scriptPath, "elapsed", elapsed,
	)
}

// toUpper 将 key 转为大写, 替换非法环境变量字符。
func toUpper(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			b = append(b, c-32) // to upper
		} else if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			b = append(b, c)
		} else {
			b = append(b, '_')
		}
	}
	return string(b)
}
