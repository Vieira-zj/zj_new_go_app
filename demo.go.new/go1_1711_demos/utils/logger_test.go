package utils

import (
	"testing"

	"go.uber.org/zap"
)

func TestLoggerPrintString(t *testing.T) {
	path := "/tmp/test/app.log"
	InitLogger(path)
	Debug("it's a debug message", zap.String("env", "dev"))
	Info("it's a info message", zap.String("env", "test"))
	Warn("it's a warn message", zap.String("env", "live"))
}

func TestLoggerPrintStruct(t *testing.T) {
	path := "/tmp/test/app.log"
	InitLogger(path)

	user := struct {
		Name string `json:"name"`
	}{
		Name: "foo",
	}
	Info("log test for struct", zap.Any("user", user))
}
