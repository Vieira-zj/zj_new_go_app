package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
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

func TrackTime() func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		fmt.Printf("elapsed: %.2fs\n", elapsed.Seconds())
	}
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
