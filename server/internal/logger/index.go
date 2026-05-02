package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.Logger

// SugaredLogger 全局 SugaredLogger，任意位置可直接使用
var SugaredLogger *zap.SugaredLogger

// 初始化全局日志（控制台 + application.log 轮转文件）
func init() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stack",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 日志轮转配置
	lumberjackLogger := &lumberjack.Logger{
		Filename:   "application.log",
		MaxSize:    100,  // 每个日志文件最大 100MB
		MaxBackups: 5,    // 最多保留 5 个旧日志文件
		MaxAge:     30,   // 日志文件最多保留 30 天
		Compress:   true, // 压缩旧日志
	}

	// 同时输出到控制台和文件
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleWriter := zapcore.AddSync(os.Stdout)
	fileWriter := zapcore.AddSync(lumberjackLogger)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleWriter, zap.InfoLevel),
		zapcore.NewCore(consoleEncoder, fileWriter, zap.InfoLevel),
	)

	globalLogger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	SugaredLogger = globalLogger.Sugar()
	zap.ReplaceGlobals(globalLogger)
}
