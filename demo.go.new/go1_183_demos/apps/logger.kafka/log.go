package main

import (
	"io"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	l   *zap.Logger
	cfg zap.Config
	// 这里使用zap.AtomicLevel类型存储logger的level信息，基于zap.AtomicLevel的level支持热更新，我们可以在程序运行时动态修改logger的log level
	level zap.AtomicLevel
}

func NewLogger(writer io.Writer, level int8, opts ...zap.Option) *Logger {
	if writer == nil {
		panic("the writer is nil")
	}

	atomicLevel := zap.NewAtomicLevelAt(zapcore.Level(level))
	logger := &Logger{
		cfg:   zap.NewProductionConfig(),
		level: atomicLevel,
	}

	logger.cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339)) // 2021-11-19 10:11:30.777
	}
	logger.cfg.EncoderConfig.TimeKey = "logtime"

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(logger.cfg.EncoderConfig),
		zapcore.AddSync(writer),
		atomicLevel,
	)
	logger.l = zap.New(core, opts...)
	return logger
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.l.Info(msg, fields...)
}

func (l *Logger) Sync() {
	l.l.Sync()
}

// SetLevel alters the logging level on runtime. it is concurrent-safe.
func (l *Logger) SetLevel(level int8) error {
	l.level.SetLevel(zapcore.Level(level))
	return nil
}
