package setup

import (
	"os"
	"path/filepath"
	"time"

	"micro-net-hub/internal/config"
	"micro-net-hub/internal/global"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// InitLogger 初始化日志（zap 输出，lumberjack 负责切割）。
//
//   - <logs.path>/info.log：logs.level 及以上、Error 以下的日志
//   - <logs.path>/error.log：Error 及以上的日志
//   - 两份同时输出到 stdout
//
// 每天 0 点（本地时间）和文件超过 logs.max-size 时切割，切下来的文件名形如 info-2026-11-22T00-00-00.000.log
// （时间为切割时刻，本地时间），按 logs.max-backups、logs.max-age 清理，logs.compress 为 true 时压缩。
// logs.level 修改后热加载生效。
func InitLogger() {
	logsConf := config.Conf.Logs
	global.LogLevel.SetLevel(logsConf.Level)

	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:    "msg",
		LevelKey:      "level",
		TimeKey:       "time",
		NameKey:       "name",
		CallerKey:     "file",
		FunctionKey:   "func",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	newFile := func(name string) *lumberjack.Logger {
		return &lumberjack.Logger{
			Filename:   filepath.Join(logsConf.Path, name), // 目录不存在时自动创建
			MaxSize:    logsConf.MaxSize,                   // 单位 MB
			MaxAge:     logsConf.MaxAge,                    // 切下来的文件保留天数
			MaxBackups: logsConf.MaxBackups,                // 切下来的文件最多保留个数
			LocalTime:  true,                               // 切下来的文件名用本地时间
			Compress:   logsConf.Compress,
		}
	}
	infoFile, errorFile := newFile("info.log"), newFile("error.log")

	// info.log：配置的级别及以上、Error 以下；error.log：Error 及以上（配置的级别高于 Error 时以配置为准）
	infoLevel := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return global.LogLevel.Enabled(l) && l < zapcore.ErrorLevel
	})
	errorLevel := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return global.LogLevel.Enabled(l) && l >= zapcore.ErrorLevel
	})
	stdout := zapcore.AddSync(os.Stdout)
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(infoFile), stdout), infoLevel),
		zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(zapcore.AddSync(errorFile), stdout), errorLevel),
	)

	logger := zap.New(core, zap.AddCaller())
	global.BasicLog = logger
	global.Log = logger.Sugar()

	go rotateDaily(infoFile, errorFile)
	global.Log.Infof("初始化zap日志完成! 日志目录: %s，级别: %s", logsConf.Path, global.LogLevel.Level())
}

// rotateDaily 每天 0 点（本地时间）切割日志；空文件不切，避免每天产生空的备份
func rotateDaily(files ...*lumberjack.Logger) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		time.Sleep(time.Until(next))
		rotateNonEmpty(files...)
	}
}

func rotateNonEmpty(files ...*lumberjack.Logger) {
	for _, f := range files {
		if fi, err := os.Stat(f.Filename); err == nil && fi.Size() > 0 {
			_ = f.Rotate()
		}
	}
}
