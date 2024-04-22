package demos_test

import (
	"errors"
	"fmt"
	"go/format"
	"log/slog"
	"math/rand"
	"os"
	"testing"
	"time"
	"unicode"
)

// Demo: Go Built-in Modules

func TestUnicode(t *testing.T) {
	t.Log("IsDigit:", unicode.IsDigit(rune('3')))
	t.Log("IsDigit:", unicode.IsDigit(rune('a')))

	t.Log("IsLower:", unicode.IsLower(rune('b')))
	t.Log("IsLower:", unicode.IsLower(rune('B')))
}

func TestRandom(t *testing.T) {
	t.Run("rand directly", func(t *testing.T) {
		for i := 0; i < 6; i++ {
			num := rand.Intn(10)
			t.Log("rand number:", num)
		}
	})

	t.Run("new rand", func(t *testing.T) {
		rander := rand.New(rand.NewSource(time.Now().Unix()))
		for i := 0; i < 6; i++ {
			num := rander.Intn(10)
			t.Log("rand number:", num)
		}
	})
}

func TestCalTime(t *testing.T) {
	duration, err := time.ParseDuration("5m")
	if err != nil {
		t.Fatal(err)
	}

	ti := time.Now().Add(duration)
	t.Log("now after 5m:", ti)

	ti = ti.AddDate(0, 0, 6)
	t.Log("now after 3 days:", ti)
}

func TestErrors(t *testing.T) {
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

func TestGoSlog(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	logger.Debug("text debug level log", "uid", 1002)
	logger.Info("text info level log", "uid", 1002)

	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	jsonLogger.Debug("json info level log", "uid", 1002)
	jsonLogger.Info("json info level log", "uid", 1002)
}

func TestGoFormat(t *testing.T) {
	b := []byte(`
	package main
	import  "fmt"


	func main(){
	  fmt.Println("hello");fmt.Println("world")
	}
`)

	// it will not check go compile error
	fb, err := format.Source(b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("format go:\n" + string(fb))
}
