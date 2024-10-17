package kafkalogger

import (
	"io"

	"go.uber.org/zap/zapcore"
)

func NewFileSyncer(writer io.Writer) zapcore.WriteSyncer {
	if ws, ok := writer.(zapcore.WriteSyncer); ok {
		return ws
	}
	// zapcore.Lock 用于将一个普通的 zapcore.WriteSyncer 包装成并发安全的 WriteSyncer
	return zapcore.Lock(zapcore.AddSync(writer))
}
