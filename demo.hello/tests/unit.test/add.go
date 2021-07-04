package unittest

import (
	"context"
	"errors"
	"strconv"
	"sync/atomic"
)

// Add .
func Add(a, b int) int {
	return a + b
}

// Counter .
type Counter int32

// IncrV1 非并发安全
func (c *Counter) IncrV1() {
	*c++
}

// IncrV2 并发安全
func (c *Counter) IncrV2() {
	atomic.AddInt32((*int32)(c), 1)
}

// I .
type I interface {
	Foo() error
}

type impl string

func (i impl) Foo() error {
	return errors.New(string(i))
}

// Bar .
func Bar(i1, i2 I) error {
	i1.Foo()
	return i2.Foo()
}

// IFoo .
type IFoo interface {
	Foo(ctx context.Context, i int) (int, error)
}

type bar struct {
	i IFoo
}

func (b bar) BarV1(ctx context.Context, i int) (int, error) {
	i, err := b.i.Foo(context.Background(), i)
	return i + 1, err
}

func (b bar) BarV2(ctx context.Context, i int) (int, error) {
	ctx = context.WithValue(ctx, impl("k"), strconv.Itoa(i))
	i, err := b.i.Foo(ctx, i)
	return i + 1, err
}
