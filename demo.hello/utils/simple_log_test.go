package utils

import "testing"

func TestLoggerPrint(t *testing.T) {
	logger := NewSimpleLog()
	logger.Debug("debug test.")
	logger.Info("info test.")
	logger.Warning("warnning test.")
	logger.Error("error test.")
	logger.Fatal("fatal test.")
}

func TestLoggerPrintWithLevel(t *testing.T) {
	logger := NewSimpleLog()
	logger.SetLevel("info")
	logger.Debug("debug test.")
	logger.Info("info test.")
	logger.Warning("warnning test.")
}

func TestLoggerPrintWithFile(t *testing.T) {
	logger := NewSimpleLog()
	if err := logger.AddFileHandler("/tmp/test/test.txt"); err != nil {
		t.Fatal(err)
	}
	logger.Debug("debug test, and output to file.")
	logger.Info("info test, and output to file.")
}
