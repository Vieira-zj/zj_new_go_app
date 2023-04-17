package kafkalogger

import (
	"io"

	"go.uber.org/zap/zapcore"
)

func NewFileSyncer(writer io.Writer) zapcore.WriteSyncer {
	if ws, ok := writer.(zapcore.WriteSyncer); ok {
		return ws
	}
	// zapcore.Lock用于将一个普通的zapcore.WriteSyncer包装成并发安全的WriteSyncer
	return zapcore.Lock(zapcore.AddSync(writer))
}
