package algorithm

import (
	"context"
	"fmt"
)

type Interceptor interface {
	intercept(*MyInterceptorChain) error
}

// InterceptorChain

type MyInterceptorChain struct {
	index        int
	ctx          context.Context
	interceptors []Interceptor
	errors       []error
}

func NewMyInterceptorChain(index int, interceptors []Interceptor) *MyInterceptorChain {
	return &MyInterceptorChain{
		index:        index,
		ctx:          context.Background(),
		interceptors: interceptors,
		errors:       make([]error, 0),
	}
}

func (chain *MyInterceptorChain) Proceed() {
	if chain.index >= len(chain.interceptors) {
		fmt.Println("end of chain")
		return
	}

	interceptor := chain.interceptors[chain.index]
	chain.index += 1
	next := chain
	if err := interceptor.intercept(next); err != nil {
		chain.errors = append(chain.errors, err)
	}
}

// Interceptors

type KeyName string

type MyInterceptorOne struct {
	name string
}

func (i *MyInterceptorOne) intercept(next *MyInterceptorChain) error {
	fmt.Println("MyInterceptorOne start:", i.name)
	key := KeyName("name")
	next.ctx = context.WithValue(next.ctx, key, "foo")

	next.Proceed()

	result := next.ctx.Value(key).(string)
	fmt.Println("response:", result)

	fmt.Println("MyInterceptorOne end:", i.name)
	return nil
}

type MyInterceptorTwo struct {
	name string
}

func (i *MyInterceptorTwo) intercept(next *MyInterceptorChain) error {
	fmt.Println("MyInterceptorTwo start:", i.name)
	key := KeyName("name")
	name := next.ctx.Value(key).(string)
	name = "hello " + name
	next.ctx = context.WithValue(next.ctx, key, name)

	next.Proceed()

	fmt.Println("MyInterceptorTwo end:", i.name)
	return nil
}

type MyInterceptorErr struct {
	name    string
	isError bool
}

func (i *MyInterceptorErr) intercept(next *MyInterceptorChain) error {
	fmt.Println("MyInterceptorErr start:", i.name)
	if i.isError {
		return fmt.Errorf("mock error")
	}

	next.Proceed()

	fmt.Println("MyInterceptorErr end:", i.name)
	return nil
}
