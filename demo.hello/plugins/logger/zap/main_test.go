package main

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestExample01(t *testing.T) {
	logger := zap.NewExample()
	defer logger.Sync()

	url := "www.google.com"
	sugar := logger.Sugar()
	sugar.Infow("failed to fetch URL",
		"url", url,
		"attempt", 3,
		"backoff", time.Second,
	)

	logger.Info("failed to fetch URL",
		zap.String("url", url),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)

	sugar.Infof("Failed to fetch URL: %s", url)

}

func TestExample02(t *testing.T) {
	// use namespace
	logger := zap.NewExample()
	defer logger.Sync()

	logger.Info("tracked some metrics",
		zap.Namespace("metrics"),
		zap.Int("counter", 1),
	)

	logger2 := logger.With(
		zap.Namespace("metrics"),
		zap.Int("counter", 1),
	)
	logger2.Info("tracked some more metrics")
}

func TestExample03(t *testing.T) {
	// use options
	logger, err := zap.NewProduction(zap.AddCaller())
	if err != nil {
		t.Fatal(err)
	}

	logger.Info("hello world")
}
