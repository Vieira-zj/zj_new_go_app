package funcproxy

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	Impls      map[string]any
	ProxyImpls map[string]any
)

func RegisterImpl(name string, impl any) {
	if Impls == nil {
		Impls = make(map[string]any, 2)
	}
	Impls[name] = impl
}

func RegisterProxyImpl(name string, impl_ any) {
	if ProxyImpls == nil {
		ProxyImpls = make(map[string]any, 2)
	}
	ProxyImpls[name] = impl_
}

func Inject() error {
	for _, implPtr := range Impls {
		valueOf := reflect.ValueOf(implPtr)
		if valueOf.Kind() != reflect.Ptr {
			return fmt.Errorf("not struct pointer")
		}

		valueOfElem := valueOf.Elem()
		typeOf := valueOfElem.Type()
		if typeOf.Kind() != reflect.Struct {
			return fmt.Errorf("not struct")
		}

		for i := 0; i < valueOfElem.NumField(); i++ {
			field := typeOf.Field(i)
			if val, ok := field.Tag.Lookup("autowire"); ok {
				fmt.Println("debug: autowire=" + val)
				implInst := valueOfElem.Field(i)
				if implInst.IsValid() && implInst.CanSet() {
					rawPtr := Impls[val]
					proxyPtr := ProxyImpls[val]
					if err := ProxyFunc(rawPtr, proxyPtr); err != nil {
						return err
					}
					// inject Impl by Impl_ proxy instance
					implInst.Set(reflect.ValueOf(proxyPtr))
				}
			}
		}
	}

	return nil
}

func ProxyFunc(rawPtr, proxyPtr any) error {
	valueOfProxy := reflect.ValueOf(proxyPtr)
	valueOfProxyElem := valueOfProxy.Elem()
	typeOfProxyElem := valueOfProxyElem.Type()
	if valueOfProxyElem.Kind() != reflect.Struct {
		return fmt.Errorf("invalid struct ptr %+v", proxyPtr)
	}

	valueOfRaw := reflect.ValueOf(rawPtr)
	valueOfRawElem := valueOfRaw.Elem()

	for i := 0; i < valueOfProxyElem.NumField(); i++ {
		fnType := typeOfProxyElem.Field(i)
		rawFnName := strings.TrimSuffix(fnType.Name, "_") // SayHello_ -> SayHello
		rawFn := valueOfRaw.MethodByName(rawFnName)
		if !rawFn.IsValid() {
			rawFn = valueOfRawElem.FieldByName(rawFnName)
		}

		f := valueOfProxyElem.Field(i)
		if f.Kind() == reflect.Func && f.IsValid() && f.CanSet() {
			// replace SayHello_ with proxy fn
			f.Set(reflect.MakeFunc(fnType.Type, makeProxyFunc(rawFnName, rawFn)))
		}
	}

	return nil
}

type ReflectFn func(in []reflect.Value) []reflect.Value

// makeProxyFunc makes proxy function with before and after interceptor.
func makeProxyFunc(name string, fn reflect.Value) ReflectFn {
	return func(in []reflect.Value) []reflect.Value {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("[aop proxy] get error:", r)
			}
		}()

		ctx := &InvocationContext{
			MethodName: name,
			Params:     in,
		}

		interceptor := InvocationInterceptor{}
		interceptor.BeforeInvoke(ctx)

		out := fn.Call(in)

		ctx.ReturnValues = out
		interceptor.AfterInvoke(ctx)

		return out
	}
}
