package utils_test

import (
	"context"
	"testing"
	"time"
)

func TestCtxWithoutCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	asyncCtx := context.WithoutCancel(ctx)
	go func() {
		for range 3 {
			select {
			case <-asyncCtx.Done():
				t.Log("goroutine exit by canceled")
				return
			default:
				t.Log("goroutine processing ...")
				time.Sleep(time.Second)
			}
		}
	}()

	time.Sleep(time.Second)
	cancel()
	t.Log("cancel main context")

	time.Sleep(3 * time.Second)
	t.Log("finished")
}
