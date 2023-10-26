package utils

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
)

// built in js fn

const scriptTitleFn = `
function title(value) {
	return value[0].toUpperCase() + value.substr(1, value.length)
}
`

type JsEngine struct {
	timeout time.Duration
	vm      *goja.Runtime
}

func NewJsEngine() JsEngine {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	for _, script := range []string{
		scriptTitleFn,
	} {
		vm.RunString(script)
	}

	return JsEngine{
		timeout: 3 * time.Second,
		vm:      vm,
	}
}

func (e JsEngine) RunBoolExp(exp string) (bool, error) {
	script := fmt.Sprintf("Boolean(%s)", exp)
	res, err := e.Run(script)
	if res == nil {
		return false, fmt.Errorf("invalid bool exp: %s", exp)
	}
	return res.(bool), err
}

func (e JsEngine) Run(script string) (any, error) {
	res, err := e.vm.RunString(script)
	return res.Export(), err
}

// Call calls built-in js function.
func (e JsEngine) Call(fnName string, params ...any) (any, error) {
	time.AfterFunc(e.timeout, func() {
		e.vm.Interrupt("execution timeout")
	})

	fn, ok := goja.AssertFunction(e.vm.Get(fnName))
	if !ok {
		return nil, fmt.Errorf("not a function: %s", fnName)
	}

	var values []goja.Value
	for _, param := range params {
		values = append(values, e.vm.ToValue(param))
	}

	res, err := fn(goja.Undefined(), values...)
	if err != nil {
		return nil, err
	}
	return res.Export(), nil
}

// GetJsonValue returns json node value for specific path.
func (e JsEngine) GetJsonValue(s any, path string) (any, error) {
	e.vm.Set("s", s)
	res, err := e.vm.RunString("s." + path)
	if res == nil {
		return "", fmt.Errorf("value not found for path: %s", path)
	}
	return res.Export(), err
}
