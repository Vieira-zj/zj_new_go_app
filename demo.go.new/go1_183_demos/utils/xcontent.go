package utils

import (
	"context"
	"time"
)

// refer: https://github.com/golang/tools/blob/master/internal/xcontext/xcontext.go

type xContent struct {
	parent context.Context
}

// Detach returns a context that keeps all the values of its parent context
// but detaches from the cancellation and error handling.
func Detach(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.TODO()
	}
	return &xContent{parent: ctx}
}

func (c xContent) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (c xContent) Done() <-chan struct{} {
	return nil
}

func (c xContent) Err() error {
	return nil
}

func (c xContent) Value(key any) any {
	return c.parent.Value(key)
}
