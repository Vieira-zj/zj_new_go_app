package utils_test

import (
	"log/slog"
	"testing"

	"demo.apps/utils"
)

func TestCustomLogger(t *testing.T) {
	l := utils.NewLogger(utils.LevelDebug)
	l.Debug("custom debug message", slog.String("hello", "world"))
	l.Trace("custom trace message", slog.String("hello", "world"))
	l.Info("custom info message", slog.String("hello", "world"))

	t.Log("set logger level to info")
	l.SetLevel(utils.LevelInfo)
	l.Debug("custom debug message", slog.String("hello", "world"))
	l.Trace("custom trace message", slog.String("hello", "world"))
	l.Info("custom info message", slog.String("hello", "world"))
}
