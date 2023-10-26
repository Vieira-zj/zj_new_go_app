package utils_test

import (
	"encoding/json"
	"testing"
	"time"

	"demo.apps/utils"
	"github.com/dop251/goja"
)

func TestRunJsExp(t *testing.T) {
	vm := goja.New()
	v, err := vm.RunString("1 + 2")
	if err != nil {
		t.Fatal(err)
	}

	num := v.Export().(int64)
	t.Logf("result: %d", num)
}

func TestRunJsFunc(t *testing.T) {
	const script = `
function sum(a, b) {
	return +a + b
}
`
	vm := goja.New()
	time.AfterFunc(200*time.Millisecond, func() {
		vm.Interrupt("halt by timeout")
	})

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatal(err)
	}

	sum, ok := goja.AssertFunction(vm.Get("sum"))
	if !ok {
		t.Fatal("not a function")
	}

	res, err := sum(goja.Undefined(), vm.ToValue(20), vm.ToValue(1))
	if err != nil {
		t.Fatal(err)
	}

	t.Log("result:", res.String())
}

func TestRunJsStruct(t *testing.T) {
	type S struct {
		Field int `json:"field"`
	}

	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	vm.Set("s", S{Field: 12})
	res, err := vm.RunString("s.field")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("result:", res.Export())
}

// Test: JsEngine

func TestJsEngineRunBoolExp(t *testing.T) {
	engine := utils.NewJsEngine()
	for _, exp := range []string{
		"2 > 1",
		"3 > 10",
		`"abc"`,
	} {
		res, err := engine.RunBoolExp(exp)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("exp:%s, result:%v", exp, res)
	}
}

func TestJsEngineCallFn(t *testing.T) {
	engine := utils.NewJsEngine()
	res, err := engine.Call("title", "abc")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("result:", res)
}

func TestJsEngineGetJsonValue(t *testing.T) {
	type S struct {
		A struct {
			B string `json:"b"`
			C struct {
				ID int `json:"id"`
			} `json:"c"`
		} `json:"a"`
	}

	s := S{}
	b := []byte(`{"a":{"b":"test","c":{"id":1010}}}`)
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}

	engine := utils.NewJsEngine()
	for _, path := range []string{
		"a.b",
		"a.c.id",
	} {
		res, err := engine.GetJsonValue(s, path)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("path=%s, result:%v", path, res)
	}
}
