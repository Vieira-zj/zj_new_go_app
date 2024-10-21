package utils_test

import (
	"log/slog"
	"sync"
	"testing"
	"time"

	"demo.apps/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func TestZapLoggerWithCallerInfo(t *testing.T) {
	cfg := zap.NewDevelopmentConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	logger, err := cfg.Build()
	assert.NoError(t, err)
	defer logger.Sync()

	// with caller info
	logger = logger.WithOptions(zap.AddCaller())
	logger.Info("failed to fetch URL",
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		// with name
		llogger := logger.Named("worker1")
		for i := 0; i < 5; i++ {
			time.Sleep(time.Millisecond * 200)
			llogger.Info("from worker 1", zap.Int("idx", i))
		}
	}()

	go func() {
		defer wg.Done()
		// with name and key
		llogger := logger.Named("worker2")
		llogger = llogger.With(zap.String("duplicate-key", "duplicate-value"))
		for i := 0; i < 5; i++ {
			time.Sleep(time.Millisecond * 200)
			llogger.Info("from worker 2", zap.Int("idx", i))
		}
	}()

	wg.Wait()
	t.Log("log finish")
}
