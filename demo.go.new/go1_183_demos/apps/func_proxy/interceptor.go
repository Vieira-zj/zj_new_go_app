package funcproxy

import (
	"fmt"
	"reflect"
)

type InvocationContext struct {
	MethodName   string
	Params       []reflect.Value
	ReturnValues []reflect.Value
}

type InvocationInterceptor struct{}

func (i InvocationInterceptor) BeforeInvoke(ctx *InvocationContext) {
	fmt.Printf("BeforeInvoke: method=%s\n", ctx.MethodName)
	fmt.Println("input params:")
	for _, p := range ctx.Params {
		fmt.Println("\t", p)
	}
}

func (i InvocationInterceptor) AfterInvoke(ctx *InvocationContext) {
	fmt.Printf("AfterInvoke: method=%s\n", ctx.MethodName)
	fmt.Println("return value:")
	for _, val := range ctx.ReturnValues {
		fmt.Println("\t", val)
	}
}
