package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

var (
	logLevelFlagsText  = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	defaultCallerDepth = 2
)

// LogLevel .
type LogLevel int

const (
	// DEBUG .
	DEBUG LogLevel = iota
	// INFO .
	INFO
	// WARNING .
	WARNING
	// ERROR .
	ERROR
	// FATAL .
	FATAL
)

// SimpleLog .
type SimpleLog struct {
	logger        *log.Logger
	logLevel      LogLevel
	isMultiWriter bool
}

/*
log.SetFlags(log.LstdFlags | log.Lshortfile)

log flags:
const (
    Ldate         = 1 << iota     // 日期示例 2009/01/23
    Ltime                         // 时间示例 01:23:23
    Lmicroseconds                 // 毫秒示例 01:23:23.123123

	Llongfile                     // 绝对路径和行号 /a/b/c/d.go:23
    Lshortfile                    // 文件和行号 d.go:23

	LUTC                          // 日期时间转为0时区的
    LstdFlags     = Ldate | Ltime // Go提供的标准抬头信息
)
*/

// NewSimpleLog .
func NewSimpleLog() *SimpleLog {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	return &SimpleLog{
		logLevel: DEBUG,
		logger:   logger,
	}
}

// SetLevel .
func (slog *SimpleLog) SetLevel(logLevel string) {
	logLevel = strings.ToUpper(logLevel)
	for idx, name := range logLevelFlagsText {
		if name == logLevel {
			slog.logLevel = LogLevel(idx)
			return
		}
	}
}

// AddFileHandler .
func (slog *SimpleLog) AddFileHandler(logPath string) error {
	if !HasPermission(logPath) {
		return fmt.Errorf("no permission: %s", logPath)
	}

	if !IsExist(logPath) {
		dirPath := filepath.Dir(logPath)
		if err := MakeDir(dirPath); err != nil {
			return fmt.Errorf("make dir [%s] error: %v", dirPath, err)
		}
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open log file [%s] error: %v", logPath, err)
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	slog.logger.SetOutput(multiWriter)
	slog.isMultiWriter = true
	return nil
}

// Debug .
func (slog *SimpleLog) Debug(v ...interface{}) {
	if slog.logLevel <= DEBUG {
		slog.setPrefix(DEBUG)
		slog.logger.Println(v...)
	}
}

// Info .
func (slog *SimpleLog) Info(v ...interface{}) {
	if slog.logLevel <= INFO {
		slog.setPrefix(INFO)
		slog.logger.Println(v...)
	}
}

// Warning .
func (slog *SimpleLog) Warning(v ...interface{}) {
	if slog.logLevel <= WARNING {
		slog.setPrefix(WARNING)
		slog.logger.Println(v...)
	}
}

// Error .
func (slog *SimpleLog) Error(v ...interface{}) {
	if slog.logLevel <= ERROR {
		slog.setPrefix(ERROR)
		slog.logger.Println(v...)
	}
}

// Fatal .
func (slog *SimpleLog) Fatal(v ...interface{}) {
	slog.setPrefix(FATAL)
	slog.logger.Println(v...)
}

func (slog *SimpleLog) setPrefix(logLevel LogLevel) {
	logPrefix := fmt.Sprintf("[%s] ", slog.getLogLevelFlagText(logLevel))
	_, file, line, ok := runtime.Caller(defaultCallerDepth)
	if ok {
		logPrefix = fmt.Sprintf("[%s][%s:%d] ", slog.getLogLevelFlagText(logLevel), filepath.Base(file), line)
	}
	slog.logger.SetPrefix(logPrefix)
}

func (slog *SimpleLog) getLogLevelFlagText(logLevel LogLevel) string {
	text := logLevelFlagsText[logLevel]
	if slog.isMultiWriter {
		return text
	}

	// only print color text to stdout
	switch logLevel {
	case INFO:
		return color.BlueString(text)
	case WARNING:
		return color.YellowString(text)
	case ERROR, FATAL:
		return color.RedString(text)
	default:
		return text
	}
}
