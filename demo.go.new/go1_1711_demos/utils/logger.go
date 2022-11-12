package utils

import (
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger     *zap.Logger
	loggerOnce sync.Once
)

func InitLogger(path string) {
	loggerOnce.Do(func() {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 设置时间格式
		encoder := zapcore.NewJSONEncoder(encoderConfig)      // 输出json格式

		fileWriteSyncer, err := getFileLogWriter(path)
		if err != nil {
			panic(err)
		}

		var core zapcore.Core
		if isProduct() {
			core = zapcore.NewCore(encoder, fileWriteSyncer, zapcore.InfoLevel)
		} else {
			// 同时写入console和文件
			core = zapcore.NewTee(
				zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
				zapcore.NewCore(encoder, fileWriteSyncer, zapcore.DebugLevel),
			)
		}
		logger = zap.New(core)
	})
}

func GetLogger() *zap.Logger {
	return logger
}

func isProduct() bool {
	env, ok := os.LookupEnv("ENV")
	if !ok {
		return false
	}
	return strings.EqualFold(env, "live")
}

// log file writer

func getFileLogWriter(path string) (zapcore.WriteSyncer, error) {
	if !isProduct() {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		return zapcore.AddSync(file), nil
	}

	return getRotatedFileLogWriter(path), nil
}

func getRotatedFileLogWriter(path string) (writeSyncer zapcore.WriteSyncer) {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    100, // 单个文件最大 100M
		MaxBackups: 60,  // 多于 60 个日志文件后，清理较旧的日志
		MaxAge:     1,   // 一天一切割
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// print log

func Debug(message string, fields ...zap.Field) {
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	logger.Debug(message, fields...)
}

func Info(message string, fields ...zap.Field) {
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	logger.Info(message, fields...)
}

func Warn(message string, fields ...zap.Field) {
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	logger.Warn(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	callerFields := getCallerInfoForLog()
	fields = append(fields, callerFields...)
	logger.Error(message, fields...)
}

func getCallerInfoForLog() []zap.Field {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return nil
	}

	funcName := runtime.FuncForPC(pc).Name()
	funcName = path.Base(funcName)
	return []zap.Field{
		zap.String("func", funcName),
		zap.String("file", file),
		zap.Int("line", line),
	}
}
