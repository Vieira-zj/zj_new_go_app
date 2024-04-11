package chain

import (
	"context"
	"errors"
	"log"
	"testing"
)

type KeyName string

type MyInterceptorOne struct {
	name string
}

func (i *MyInterceptorOne) apply(chain *InterceptorsChain) error {
	log.Println("start:", i.name)
	chain.ctx = context.WithValue(chain.ctx, KeyName("text"), "foo")

	chain.Next()

	result, ok := chain.ctx.Value(KeyName("newtext")).(string)
	if !ok {
		return errors.New("key [new-text] not found in context")
	}
	log.Println("result:", result)
	log.Println("end:", i.name)
	return nil
}

type MyInterceptorTwo struct {
	name string
}

func (i *MyInterceptorTwo) apply(chain *InterceptorsChain) error {
	log.Println("start:", i.name)
	name := chain.ctx.Value(KeyName("text")).(string)
	name = "hello, " + name
	chain.ctx = context.WithValue(chain.ctx, KeyName("newtext"), name)

	chain.Next()

	log.Println("end:", i.name)
	return nil
}

type MyInterceptorErr struct {
	name    string
	isError bool
}

func (i *MyInterceptorErr) apply(chain *InterceptorsChain) error {
	log.Println("start:", i.name)
	if i.isError {
		return errors.New("mock error")
	}

	chain.Next()

	log.Println("end:", i.name)
	return nil
}

func TestInterceptorsChain(t *testing.T) {
	interceptors := []Interceptor{
		&MyInterceptorOne{
			name: "1st-interceptor",
		},
		&MyInterceptorErr{
			name:    "err-interceptor",
			isError: false,
		},
		&MyInterceptorTwo{
			name: "2nd-interceptor",
		},
	}

	chain := NewInterceptorsChain(0, interceptors)
	chain.Next()

	for _, err := range chain.errors {
		t.Log("chain error:", err)
	}
	t.Log("run chain finish")
}
