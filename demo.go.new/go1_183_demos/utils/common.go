package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
)

// BinaryCeil rounds up the given uint32 value to the nearest power of 2.
func BinaryCeil(v uint32) uint32 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v++
	return v
}

func ConcatString(sl ...string) string {
	n := 0
	for i := 0; i < len(sl); i++ {
		n += len(sl[i])
	}

	b := strings.Builder{}
	b.Grow(n)
	for _, s := range sl {
		b.WriteString(s)
	}
	return b.String()
}

func JsonMarshalStream(r io.Reader, object any) error {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	return decoder.Decode(object)
}

func JsonUnmarshalStream(w io.Writer, object any) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(object)
}

func GetSlogLevel() slog.Level {
	var curLevel slog.Level = -10
	for _, level := range []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		if enabled := slog.Default().Enabled(context.TODO(), level); enabled {
			curLevel = level
			break
		}
	}
	return curLevel
}

func IsNil(x any) bool {
	if x == nil {
		return true
	}
	return reflect.ValueOf(x).IsNil()
}

func IsEmptyStruct(x any) bool {
	valOf := reflect.ValueOf(x)
	if valOf.Kind() == reflect.Ptr {
		valOf = valOf.Elem()
	}
	return valOf.IsZero()
}

func TrackTime() func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		fmt.Printf("elapsed: %.2fs\n", elapsed.Seconds())
	}
}

func DelFirstNItemsOfSlice(s []any /* will change input slice */, n int) ([]any, error) {
	if n >= len(s) {
		return nil, fmt.Errorf("n must be less than length of input slice")
	}

	m := copy(s, s[n:])
	for i := m; i < len(s); i++ {
		s[i] = nil // avoid memory leaks
	}

	s = s[:m] // reset length
	return s, nil
}

func DeepCopy(src, dest any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(b, &dest); err != nil {
		return err
	}
	return nil
}

// Panic Handler in Goroutine

func logPanic(r any) {
	b := debug.Stack()
	log.Printf("panic error: %v", r)
	log.Println("stack:\n", string(b))
}

var defaultPanicHanders = []func(any){logPanic}

func HandlePanic(handlers ...func(any)) {
	if r := recover(); r != nil {
		for _, handler := range defaultPanicHanders {
			handler(r)
		}
		for _, handler := range handlers {
			handler(r)
		}
	}
}

func Go(fn func()) {
	go func() {
		defer HandlePanic()
		fn()
	}()
}
