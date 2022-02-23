package logger

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(filename string, level string) *zap.SugaredLogger {
	// set default log file name
	if filename == "" {
		filename = "sparkle.log"
	}

	hook := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    25,
		MaxBackups: 4,
		MaxAge:     7,
		LocalTime:  true,
		Compress:   true,
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "file",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),                                          // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(hook)), // 打印到控制台和文件
		zap.NewAtomicLevelAt(func() zapcore.Level {
			if level, ok := map[string]zapcore.Level{
				"debug":  zapcore.DebugLevel,
				"info":   zapcore.InfoLevel,
				"warn":   zapcore.WarnLevel,
				"error":  zapcore.ErrorLevel,
				"dpanic": zapcore.DPanicLevel,
				"panic":  zapcore.PanicLevel,
				"fatal":  zapcore.FatalLevel,
			}[level]; ok {
				return level
			} else {
				return zapcore.ErrorLevel
			}
		}()),
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger.Sugar()
}
