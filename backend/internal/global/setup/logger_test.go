package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestInitLogger(t *testing.T) {
	dir := t.TempDir()
	oldLogs := config.Conf.Logs
	t.Cleanup(func() { config.Conf.Logs = oldLogs; global.LogLevel.SetLevel(zapcore.DebugLevel) })
	config.Conf.Logs = &config.Logs{Level: zapcore.WarnLevel, Path: dir, MaxSize: 1, MaxBackups: 2, MaxAge: 1}

	InitLogger()
	global.Log.Debug("debug-msg")
	global.Log.Info("info-msg")
	global.Log.Warn("warn-msg")
	global.Log.Error("error-msg")

	// 热加载：改成 Debug 后立即生效
	global.LogLevel.SetLevel(zapcore.DebugLevel)
	global.Log.Debug("debug-after-reload")

	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		return string(b)
	}
	info, errLog := read("info.log"), read("error.log")
	for _, c := range []struct {
		log, msg string
		want     bool
	}{
		{info, "debug-msg", false},
		{info, "info-msg", false},
		{info, "warn-msg", true},
		{info, "error-msg", false},
		{info, "debug-after-reload", true},
		{errLog, "error-msg", true},
		{errLog, "warn-msg", false},
	} {
		if got := strings.Contains(c.log, c.msg); got != c.want {
			t.Errorf("%q present=%v, want %v", c.msg, got, c.want)
		}
	}
}

func TestRotateNonEmpty(t *testing.T) {
	dir := t.TempDir()
	full := &lumberjack.Logger{Filename: filepath.Join(dir, "info.log"), LocalTime: true}
	empty := &lumberjack.Logger{Filename: filepath.Join(dir, "error.log"), LocalTime: true}
	t.Cleanup(func() { full.Close(); empty.Close() })
	if _, err := full.Write([]byte("line\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := empty.Write(nil); err != nil {
		t.Fatal(err)
	}

	rotateNonEmpty(full, empty)

	infoBackups, _ := filepath.Glob(filepath.Join(dir, "info-*.log"))
	errorBackups, _ := filepath.Glob(filepath.Join(dir, "error-*.log"))
	if len(infoBackups) != 1 || len(errorBackups) != 0 {
		t.Fatalf("backups: info=%v error=%v", infoBackups, errorBackups)
	}
	// 备份文件名用本地时间
	if want := time.Now().Format("info-2006-01-02T15"); !strings.HasPrefix(filepath.Base(infoBackups[0]), want) {
		t.Errorf("backup name %s, want prefix %s", filepath.Base(infoBackups[0]), want)
	}
}
