package utils

import (
	"context"
	"time"
)

type detachedContext struct {
	ctx context.Context
}

// Deprecated: use context.WithoutCancel() instead.
func NewDetachedContext(ctx context.Context) context.Context {
	return detachedContext{ctx}
}

func (c detachedContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c detachedContext) Done() <-chan struct{}       { return nil }
func (c detachedContext) Err() error                  { return nil }
func (c detachedContext) Value(key any) any           { return c.ctx.Value(key) }
