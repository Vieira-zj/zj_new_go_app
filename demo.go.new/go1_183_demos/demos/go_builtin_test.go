package demos_test

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"go/format"
	"io"
	"log"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode"

	"demo.apps/utils"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Demo: Go Built-in Modules

func TestBuildInFn(t *testing.T) {
	t.Run("build-in func", func(t *testing.T) {
		t.Log("min:", min(1, 2))
		t.Log("max:", max(8, 10))
	})

	t.Run("clear slice", func(t *testing.T) {
		s := []int{1, 2, 3}
		t.Logf("len=%d, cap=%d", len(s), cap(s))

		clear(s)
		t.Log("after clear")
		t.Logf("len=%d, cap=%d", len(s), cap(s))
		t.Log("slice:", s)
	})

	t.Run("clear map", func(t *testing.T) {
		m := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
		}
		t.Logf("len=%d", len(m))

		clear(m)
		t.Log("after clear")
		t.Logf("len=%d", len(m))
	})
}

func TestCompare(t *testing.T) {
	rint := cmp.Compare(2, 1)
	t.Log("result:", rint)

	rbool := cmp.Less(2, 1)
	t.Log("result:", rbool)
}

func TestMath(t *testing.T) {
	t.Run("math floor and ceil", func(t *testing.T) {
		t.Log("math floor:", math.Floor(1.1))
		t.Log("math ceil:", math.Ceil(1.1))
	})
}

func TestRandom(t *testing.T) {
	t.Run("rand directly", func(t *testing.T) {
		for i := 0; i < 6; i++ {
			num := rand.Intn(10)
			t.Log("rand number:", num)
		}
	})

	t.Run("new rand with seed", func(t *testing.T) {
		rander := rand.New(rand.NewSource(time.Now().Unix()))
		for i := 0; i < 6; i++ {
			num := rander.Intn(10)
			t.Log("rand number:", num)
		}
	})
}

func TestFmtPrint(t *testing.T) {
	s := "hello"
	fmt.Printf("type: %T\n", s)
	fmt.Printf("quota value: %q\n", s)
}

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

	t.Run("string title", func(t *testing.T) {
		c := cases.Title(language.English)
		for _, s := range []string{"task", "sub-task"} {
			t.Log("result:", c.String(s))
		}
	})

	t.Run("string prefix cut", func(t *testing.T) {
		s := "hello foo"
		result, ok := strings.CutPrefix(s, "hello ")
		assert.True(t, ok)
		t.Log("cut prefix result:", result)
	})
}

func TestSort(t *testing.T) {
	persons := []TestPerson{
		{Name: "user1", Age: 30},
		{Name: "user3", Age: 21},
		{Name: "user2", Age: 35},
	}

	t.Run("sort by int age", func(t *testing.T) {
		sort.Slice(persons, func(i, j int) bool {
			return persons[i].Age > persons[j].Age
		})
		for _, p := range persons {
			t.Logf("person: %+v", p)
		}
	})

	t.Run("sort by string name", func(t *testing.T) {
		sort.Slice(persons, func(i, j int) bool {
			return persons[i].Name < persons[j].Name
			// return strings.Compare(persons[i].Name, persons[j].Name) < 0
		})
		for _, p := range persons {
			t.Logf("person: %+v", p)
		}
	})
}

func TestDateTime(t *testing.T) {
	duration, err := time.ParseDuration("5m")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().Local()
	t.Log("now:", now.Format(time.DateTime))

	t.Run("get time hour and min", func(t *testing.T) {
		t.Logf("hour:%d, min:%d", now.Hour(), now.Minute())
	})

	t.Run("time before or after", func(t *testing.T) {
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

	t.Run("time truncate", func(t *testing.T) {
		ti := now.Truncate(24 * time.Hour)
		t.Log("truncate by day:", ti.Format(time.DateTime))

		ti = now.Truncate(time.Hour)
		t.Log("truncate by hour:", ti.Format(time.DateTime))

		ti = now.Truncate(time.Minute)
		t.Log("truncate by minute:", ti.Format(time.DateTime))
	})

	t.Run("time round", func(t *testing.T) {
		ti := now.Round(time.Hour)
		t.Log("truncate by hour:", ti.Format(time.DateTime))

		ti = now.Round(time.Minute)
		t.Log("truncate by minute:", ti.Format(time.DateTime))
	})
}

func TestTimeTicker(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	interval := 200 * time.Millisecond

	t.Run("loop by time ticker with fix interval", func(t *testing.T) {
		tick := time.NewTicker(interval)
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

	t.Run("loop by timer with diff interval", func(t *testing.T) {
		ti := time.NewTimer(interval)
		defer ti.Stop()
	outer:
		for i := 0; i < 10; i++ {
			select {
			case <-ctx.Done():
				t.Log(ctx.Err())
				break outer
			case <-ti.C:
				t.Log("tick at:", i)
				// reset to a diff interval for next iter
				ti.Reset(interval - time.Duration(i))
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

func TestHttpOps(t *testing.T) {
	t.Run("http limit reader", func(t *testing.T) {
		rb := bytes.NewBufferString("hello world")
		r := io.NopCloser(rb)

		w := httptest.NewRecorder()
		out := http.MaxBytesReader(w, r, 8)
		defer out.Close()

		outb, err := io.ReadAll(out)
		assert.NoError(t, err)
		t.Log("out bytes:", string(outb))
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

func TestErrorTypeCheck(t *testing.T) {
	for _, path := range []string{
		"/tmp/test/out.json",
		"/tmp/test/out.txt",
	} {
		b, err := os.ReadFile(path)

		switch err.(type) {
		case nil:
			t.Log("err is nil")
		case StackError:
			t.Log("stack error")
		default:
			t.Log("unexpected error")
		}

		if len(b) > 0 {
			n, err := io.Copy(io.Discard, bytes.NewReader(b))
			assert.NoError(t, err)
			t.Logf("%d bytes discard", n)
		}
	}
}

func TestGoErrors(t *testing.T) {
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

	t.Run("error wrap and unwrap", func(t *testing.T) {
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

// Demo: lo util

func TestLoSliceUnique(t *testing.T) {
	names := lo.Uniq[string]([]string{"foo", "bar", "foo"})
	t.Log("names:", names)
}

func TestLoSliceFilterAndMap(t *testing.T) {
	even := lo.Filter[int]([]int{1, 2, 3, 4}, func(x int, index int) bool {
		return x%2 == 0
	})
	t.Log("even num:", even)

	result := lo.Map[int64, string]([]int64{1, 2, 3, 4}, func(x int64, index int) string {
		return strconv.FormatInt(x, 10)
	})
	t.Log("map result:", result)
}

func TestLoMapKeysAndValues(t *testing.T) {
	m := map[string]int{"foo": 1, "bar": 2}
	keys := lo.Keys(m)
	t.Log("map keys:", keys)

	values := lo.Values(m)
	t.Log("map values:", values)

	value := lo.ValueOr(m, "test", 29)
	t.Log("value:", value)
}

func TestLoStringTest(t *testing.T) {
	sub := lo.Substring("hello", 2, 3)
	t.Log("sub str:", sub)

	str := lo.RandomString(12, lo.LettersCharset)
	t.Log("rand str:", str)
}
