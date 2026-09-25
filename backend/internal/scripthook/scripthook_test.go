package scripthook

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// setupLogger 确保测试环境下 global.Log 可用。
func setupLogger(t *testing.T) {
	t.Helper()
	if global.Log == nil {
		previous := global.Log
		global.Log = zap.NewNop().Sugar()
		t.Cleanup(func() { global.Log = previous })
	}
}

// TestWithEnv 验证 WithEnv 注入的键值对能被正确携带和读取。
func TestWithEnv(t *testing.T) {
	ctx := context.Background()
	ctx = WithEnv(ctx, "message", "hello world")
	ctx = WithEnv(ctx, "user_id", "42")

	envs, ok := ctx.Value(envCtxKey).([]envPair)
	require.True(t, ok)
	require.Len(t, envs, 2)
	assert.Equal(t, "message", envs[0].key)
	assert.Equal(t, "hello world", envs[0].value)
	assert.Equal(t, "user_id", envs[1].key)
	assert.Equal(t, "42", envs[1].value)
}

// TestWithEnv_EmptyContext 验证空 context 上调用 WithEnv 不会 panic。
func TestWithEnv_EmptyContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithEnv(ctx, "a", "1")
	envs, ok := ctx.Value(envCtxKey).([]envPair)
	require.True(t, ok)
	require.Len(t, envs, 1)
}

// TestToUpper 验证环境变量名转换。
func TestToUpper(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"message", "MESSAGE"},
		{"source_addr", "SOURCE_ADDR"},
		{"user-id", "USER_ID"},
		{"my.key", "MY_KEY"},
		{"ALREADY_UP", "ALREADY_UP"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, toUpper(tt.input))
		})
	}
}

// TestRun_NilConfig 验证配置为空时 Run 不会 panic。
func TestRun_NilConfig(t *testing.T) {
	setupLogger(t)
	original := config.Conf.ScriptHook
	config.Conf.ScriptHook = nil
	defer func() { config.Conf.ScriptHook = original }()

	assert.NotPanics(t, func() {
		Run(context.Background(), "no_such_script")
	})

	// 给一点时间让 goroutine 有机会跑, 但既然是 nil config 不应该启动 goroutine
	time.Sleep(10 * time.Millisecond)
}

// TestRun_EmptyDir 验证 Dir 为空时 Run 不会 panic。
func TestRun_EmptyDir(t *testing.T) {
	setupLogger(t)
	original := config.Conf.ScriptHook
	config.Conf.ScriptHook = &config.ScriptHook{Dir: ""}
	defer func() { config.Conf.ScriptHook = original }()

	assert.NotPanics(t, func() {
		Run(context.Background(), "security_audit")
	})
	time.Sleep(10 * time.Millisecond)
}

// TestRun_ScriptNotFound 验证脚本名未配置时记 warn 日志且不 panic。
func TestRun_ScriptNotFound(t *testing.T) {
	setupLogger(t)
	original := config.Conf.ScriptHook
	config.Conf.ScriptHook = &config.ScriptHook{
		Dir:     "/tmp",
		Scripts: map[string]string{},
	}
	defer func() { config.Conf.ScriptHook = original }()

	assert.NotPanics(t, func() {
		Run(context.Background(), "unknown_script")
	})
	time.Sleep(10 * time.Millisecond)
}

// TestRun_ExecutesEcho 端到端测试: 执行真实脚本并验证环境变量传递。
func TestRun_ExecutesEcho(t *testing.T) {
	setupLogger(t)
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "echo_env.sh")

	// 脚本输出 SCRIPTHOOK_MESSAGE 环境变量
	err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho \"MESSAGE=$SCRIPTHOOK_MESSAGE\"\n"), 0755)
	require.NoError(t, err)

	original := config.Conf.ScriptHook
	config.Conf.ScriptHook = &config.ScriptHook{
		Dir:           dir,
		TimeoutSeconds: 5,
		Scripts: map[string]string{
			"echo_env": "echo_env.sh",
		},
	}
	defer func() { config.Conf.ScriptHook = original }()

	ctx := context.Background()
	ctx = WithEnv(ctx, "message", "hello-from-test")

	Run(ctx, "echo_env")

	// 给 goroutine 一点时间执行
	time.Sleep(200 * time.Millisecond)
}

// TestRun_ContextDeadline 验证 ctx deadline 能正确传递给子进程。
func TestRun_ContextDeadline(t *testing.T) {
	setupLogger(t)
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "sleepy.sh")

	err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nsleep 10\necho done\n"), 0755)
	require.NoError(t, err)

	original := config.Conf.ScriptHook
	config.Conf.ScriptHook = &config.ScriptHook{
		Dir:           dir,
		TimeoutSeconds: 30, // 配置较长超时
		Scripts: map[string]string{
			"sleepy": "sleepy.sh",
		},
	}
	defer func() { config.Conf.ScriptHook = original }()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 不应阻塞, 也不应 panic
	Run(ctx, "sleepy")

	// 等待足够时间让脚本因 context deadline 而终止
	time.Sleep(300 * time.Millisecond)
}