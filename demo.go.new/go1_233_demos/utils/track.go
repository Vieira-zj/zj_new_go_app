package utils

import (
	"log/slog"
	"time"
)

func TrackExecTime(name string) func() {
	start := time.Now()
	return func() {
		slog.Info("exec completed",
			slog.String("name", name),
			slog.Duration("elapsed", time.Since(start)),
		)
	}
}
