package demos_test

import (
	"context"
	"errors"
	"fmt"
	"go/format"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode"

	"demo.apps/utils"
	"github.com/stretchr/testify/assert"
)

// Demo: Go Built-in Modules

func TestString(t *testing.T) {
	t.Run("char check", func(t *testing.T) {
		assert.True(t, unicode.IsDigit(rune('3')))
		assert.False(t, unicode.IsDigit(rune('a')))

		assert.True(t, unicode.IsLower(rune('b')))
		assert.False(t, unicode.IsLower(rune('B')))
	})

	t.Run("string compare", func(t *testing.T) {
		assert.True(t, strings.EqualFold("case", "CaSe"))
		assert.False(t, strings.EqualFold("case", "cases"))
	})
}

func TestRandom(t *testing.T) {
	t.Run("rand directly", func(t *testing.T) {
		for i := 0; i < 6; i++ {
			num := rand.Intn(10)
			t.Log("rand number:", num)
		}
	})

	t.Run("new rand with src", func(t *testing.T) {
		rander := rand.New(rand.NewSource(time.Now().Unix()))
		for i := 0; i < 6; i++ {
			num := rander.Intn(10)
			t.Log("rand number:", num)
		}
	})
}

func TestDateTime(t *testing.T) {
	duration, err := time.ParseDuration("5m")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()

	t.Run("get time hour and min", func(t *testing.T) {
		t.Logf("hour:%d, min:%d", now.Hour(), now.Minute())
	})

	t.Run("time before", func(t *testing.T) {
		d1, err := time.Parse(time.DateOnly, "2024-08-02")
		assert.NoError(t, err)

		d2, err := time.Parse(time.DateOnly, "2024-07-31")
		assert.NoError(t, err)

		assert.True(t, d2.Before(d1))
	})

	t.Run("time since", func(t *testing.T) {
		ti := now.Add(-duration)
		since := time.Since(ti)
		t.Logf("since: %.2f min, %.2f sec", since.Minutes(), since.Seconds())
	})

	t.Run("time calculate", func(t *testing.T) {
		ti := now.Add(duration)
		t.Log("now after 5m:", ti)

		ti = ti.AddDate(0, 0, 6)
		t.Log("now after 6 days:", ti)
	})
}

func TestTimeTicker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stop := 200 * time.Millisecond

	t.Run("loop by time ticker", func(t *testing.T) {
		tick := time.NewTicker(stop)
		defer tick.Stop()
	outer:
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				t.Log(ctx.Err())
				break outer
			case <-tick.C:
				t.Log("tick at:", i)
			}
		}
	})

	t.Run("loop by timer and reset", func(t *testing.T) {
		ti := time.NewTimer(stop)
		defer ti.Stop()
	outer:
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				t.Log(ctx.Err())
				break outer
			case <-ti.C:
				t.Log("tick at:", i)
				ti.Reset(stop)
			}
		}
	})
}

func TestFilePath(t *testing.T) {
	t.Run("get abs path", func(t *testing.T) {
		t.Log("separator:", string(filepath.Separator))

		absPath, err := filepath.Abs("./")
		assert.NoError(t, err)
		t.Log("abs path:", absPath)
	})

	t.Run("file pattern match", func(t *testing.T) {
		ok, err := filepath.Match("*_test.go", "go_test.go")
		assert.NoError(t, err)
		t.Log("is match:", ok)
	})
}

func TestGoFormat(t *testing.T) {
	b := []byte(`
	package main
	import  "fmt"


	func main(){
	  fmt.Println("hello");fmt.Println("world")
	}
`)

	// it does not check go compile error
	fb, err := format.Source(b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("format go:\n" + string(fb))
}

// Error

type StackError struct {
	err   string
	stack string
}

func (e StackError) Error() string {
	return e.err
}

func (e StackError) Stack() string {
	return e.stack
}

func TestErrors(t *testing.T) {
	t.Run("error as interface", func(t *testing.T) {
		err := StackError{
			err:   "mock error",
			stack: "mock stack",
		}
		t.Log("err:", err.Error())

		if ok := errors.As(err, new(interface{ Stack() string })); ok {
			t.Log("stack:", err.Stack())
		}
	})

	t.Run("error wrapped", func(t *testing.T) {
		err := errors.New("mock err")
		wrappedErr := fmt.Errorf("wrapped: %w", err)
		t.Log("err:", wrappedErr)

		rawErr := errors.Unwrap(wrappedErr)
		t.Log("raw err:", rawErr)
	})

	t.Run("errors join from Go 1.20", func(t *testing.T) {
		err1 := errors.New("Error 1st")
		err2 := errors.New("Error 2nd")

		err := errors.Join(err1, err2)
		t.Log("err1:", errors.Is(err, err1))
		t.Log("err2:", errors.Is(err, err2))
	})
}

// Log

func TestLogToFile(t *testing.T) {
	path := "/tmp/test/log_test.txt"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	logWriter := io.MultiWriter(f, os.Stdout)
	logger := log.New(logWriter, "[go_test]", log.LstdFlags)

	logger.Println("test log to file start")
	logger.Println("test log to file end")
	t.Log("done")
}

// Slog

//
// 原理
//
// 1. 用户调用前端 `Logger` 提供的日志记录方法 `Info` 记录一条日志
// 2. `Info` 方法会调用一个私有方法 `log`， `log` 方法内部会使用 `NewRecord` 创建一个日志条目 `Record`
// 3. 最终，`Logger` 会调用其嵌入的 `Handler` 对象的 `Handle` 方法解析 `Record` 并执行日志记录逻辑
//

func TestGoSlog(t *testing.T) {
	t.Run("log with ctx", func(t *testing.T) {
		ctx := context.TODO()
		t.Log("log level:", utils.GetSlogLevel().String())
		slog.DebugContext(ctx, "debug message", "hello", "world")
		slog.WarnContext(ctx, "warn message", "hello", "world")
	})

	t.Run("log key/value", func(t *testing.T) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		logger.Info("info message", slog.String("hello", "world"), slog.Int("code", 200), slog.Any("error", fmt.Errorf("mock err")))
	})

	t.Run("log group key/value", func(t *testing.T) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		logger.Info("info message", slog.Group("user", slog.String("name", "root"), slog.Int("age", 31)))
	})

	t.Run("with new logger", func(t *testing.T) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		l := logger.With(slog.String("trace_id", "abc-xyz"))
		l.Info("info message")
		l.Info("warn message")
	})
}

func TestSlogHandler(t *testing.T) {
	t.Run("json handler", func(t *testing.T) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:       slog.LevelDebug, // 设置日志级别
			AddSource:   true,            // 记录日志位置
			ReplaceAttr: nil,
		}))
		logger.Debug("json debug level log", "hello", "world")
		logger.Info("json info level log", "hello", "world")
	})

	t.Run("text handler", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			AddSource:   true,
			ReplaceAttr: nil,
		}))
		logger.Debug("text debug level log", "hello", "world")
		logger.Info("text info level log", "hello", "world")
	})
}

func TestSlogLogger(t *testing.T) {
	t.Run("relace default slog logger", func(t *testing.T) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   true,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
		}))
		slog.SetDefault(logger)

		slog.Info("info message", "hello", "world")
		// log is replaced too
		log.Println("normal log")
	})

	t.Run("log logger", func(t *testing.T) {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   true,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
		}))
		logLogger := slog.NewLogLogger(logger.Handler(), slog.LevelInfo)
		logLogger.Println("normal log")
	})
}
