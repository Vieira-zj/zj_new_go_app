package utils

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// Custom Logger

type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelTrace = slog.Level(-2) // 自定义 level
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

type Logger struct {
	l   *slog.Logger
	lvl *slog.LevelVar
}

func NewLogger(level Level) *Logger {
	lvl := &slog.LevelVar{}
	lvl.Set(level)

	l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     lvl,
		AddSource: true,
		// 记录日志时, 回调 ReplaceAttr 方法
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != slog.LevelKey {
				return a
			}
			level, ok := a.Value.Any().(slog.Level)
			if !ok {
				return a // not happen
			}
			if level == LevelTrace {
				a.Value = slog.StringValue("TRACE")
			}
			return a
		},
	}))

	return &Logger{l: l, lvl: lvl}
}

func (l *Logger) SetLevel(level Level) {
	l.lvl.Set(level)
}

func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.log(ctx, level, msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.Log(context.TODO(), LevelDebug, msg, args...)
}

func (l *Logger) Trace(msg string, args ...any) {
	l.Log(context.TODO(), LevelTrace, msg, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.Log(context.TODO(), LevelInfo, msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.Log(context.TODO(), LevelWarn, msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.Log(context.TODO(), LevelError, msg, args...)
}

func (l *Logger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if !l.l.Enabled(ctx, level) {
		return
	}

	var pc uintptr
	var pcs [1]uintptr
	// NOTE: 这里修改 skip 为 4, *slog.Logger.log 源码中 skip 为 3
	runtime.Callers(4, pcs[:])
	pc = pcs[0]
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	if ctx == nil {
		ctx = context.TODO()
	}
	_ = l.l.Handler().Handle(ctx, r)
}
