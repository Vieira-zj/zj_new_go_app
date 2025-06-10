package demos_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Demo: Go Test

func TestClearupCase(t *testing.T) {
	t.Cleanup(func() {
		t.Log("case clear")
	})

	if ok := false; !ok {
		t.Fatal("mock fatal")
	}
	t.Log("case run")
}

func MyStringSplit(s, sep string) []string {
	idx := strings.Index(s, sep)
	if idx == -1 {
		return []string{}
	}

	subs := make([]string, 0, 2)
	for idx != -1 {
		subs = append(subs, s[:idx])
		s = s[idx+len(sep):]
		idx = strings.Index(s, sep)
	}
	time.Sleep(time.Second)
	return subs
}

func TestParallelRunCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		sep   string
		want  []string
	}{
		{"base case", "a:b:c", ":", []string{"a", "b", "c"}},
		{"wrong sep", "a:b:c", ",", []string{"a:b:c"}},
		{"more sep", "abcd", "bc", []string{"a", "d"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MyStringSplit(tt.input, tt.sep)
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected:%#v, got:%#v", tt.want, got)
			}
		})
	}
}

// Demo: Go Stmt

type SystemType int

const (
	SysTypeWin SystemType = iota + 1
	SysTypeLinux
	SysTypeMac
)

func TestCustomConstType(t *testing.T) {
	t.Run("print custom const type", func(t *testing.T) {
		t.Log("win system:", SysTypeWin)
		t.Log("mac system:", SysTypeMac)
	})
}

func TestSwitchConds(t *testing.T) {
	getNumberDesc := func(num int) string {
		switch num {
		case 1, 2, 3:
			return "num <= 3"
		case 4, 5:
			return "3 < num <= 5"
		default:
			return "num > 5"
		}
	}

	for i := 2; i < 7; i++ {
		desc := getNumberDesc(i)
		t.Logf("%d: %s", i, desc)
	}
}

func TestTypeAssert(t *testing.T) {
	var s, m any

	t.Run("slice", func(t *testing.T) {
		s = []int{1}
		switch s.(type) {
		case []any:
			t.Log("type: []any")
		case []int:
			t.Log("type: []int")
		default:
			t.Log("unknown type")
		}
	})

	t.Run("map", func(t *testing.T) {
		m = map[string]int{"one": 1}
		switch m.(type) {
		case map[string]any:
			t.Log("type: map[string]any")
		default:
			t.Log("unknown type")
		}
	})
}

// Demo: Defer

func TestDeferFn01(t *testing.T) {
	testFn := func() func() {
		t.Log("test fn")
		return func() {
			t.Log("wrapped test fn")
		}
	}

	defer testFn()()
	t.Log("start test defer fn")
	time.Sleep(200 * time.Millisecond)
	t.Log("end test defer fn")
}

type WrappedTest struct {
	t *testing.T
}

func (w *WrappedTest) fn1() *WrappedTest {
	w.t.Log("fn1 invoke")
	return w
}

func (w *WrappedTest) fn2() *WrappedTest {
	w.t.Log("fn2 invoke")
	return w
}

func TestDeferFn02(t *testing.T) {
	s := &WrappedTest{t}
	defer s.fn1().fn2()

	t.Log("start test defer struct fn")
	time.Sleep(200 * time.Millisecond)
	t.Log("end test defer struct fn")
}

func TestDeferInBlocks(t *testing.T) {
	// no effect in code blocks
	defer func() {
		t.Log("test clearup")
	}()

	t.Log("test start")

	{
		defer func() {
			t.Log("test sub1 clearup")
		}()
		t.Log("test sub1")
	}

	{
		defer func() {
			t.Log("test sub2 clearup")
		}()
		t.Log("test sub2")
	}

	t.Log("test finish")
}

// Demo: Ref

func TestPtrInit(t *testing.T) {
	t.Run("init ptr by addr", func(t *testing.T) {
		i := 0
		var j *int //nolint:gosimple
		j = &i
		t.Log("int:", *j)

		*j = 1
		t.Log("int:", *j)
	})

	t.Run("init ptr by new", func(t *testing.T) {
		var j *int = new(int) // 为 0 分配空间
		t.Log("int:", *j)

		*j = 1
		t.Log("int:", *j)
	})
}

func TestCompareNilVal(t *testing.T) {
	var myByteNil (*byte) = nil

	t.Run("is nil", func(t *testing.T) {
		assert.True(t, myByteNil == nil)

		// error: mismatch type
		// var myIntNil (*int) = nil
		// t.Log(myByteNil == myIntNil)
	})

	t.Run("init nil", func(t *testing.T) {
		str := byte(0)
		myByteNil = &str
		t.Log("value:", *myByteNil)
	})
}

func TestRefUpdateForSlice(t *testing.T) {
	update := func(fruits []testFruit) {
		for idx := range fruits {
			fruits[idx].name += "-test"
		}
	}

	fruits := []testFruit{
		{"apple"},
		{"pair"},
	}
	t.Log("fruits:", fruits)
	update(fruits)
	t.Log("new fruits:", fruits)
}

func TestRefUpdateForMap(t *testing.T) {
	update := func(fruits map[string]testFruit) {
		for key, fruit := range fruits {
			fruit.name += "-test"
			fruits[key] = fruit
		}
	}

	fruits := map[string]testFruit{
		"apple": {"apple"},
		"pair":  {"pair"},
	}
	t.Log("fruits:", fruits)
	update(fruits)
	t.Log("new fruits:", fruits)
}

// Demo: Context

func TestContextWrapped(t *testing.T) {
	t.Run("wrapped timeout ctx", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		wrappedCtxFn := func(ctx context.Context) {
			t.Log("new local ctx with timeout")
			lctx, lcancel := context.WithTimeout(ctx, 3*time.Second)
			defer lcancel()

			<-lctx.Done()
			t.Log("cannelled")
		}

		wrappedCtxFn(ctx)
		t.Log("done")
	})

	t.Run("wrapped ctx with value", func(t *testing.T) {
		type TestCtxKey string

		const key TestCtxKey = "ctx_key"

		printCtxValue := func(ctx context.Context, key TestCtxKey) {
			if s, ok := ctx.Value(key).(string); ok {
				t.Log("ctx value:", s)
			} else {
				t.Log("ctx value not found for key:", key)
			}
		}

		ctx := context.TODO()
		ctx = context.WithValue(ctx, key, "ctx_value1")
		printCtxValue(ctx, key)

		// 由下而上, 由子及父, 依次对 key 进行匹配
		ctx = context.WithValue(ctx, key, "ctx_value2")
		printCtxValue(ctx, key)
	})
}

func TestContextKeyByStruct(t *testing.T) {
	type (
		TestCtxKey1 struct{}
		TestCtxKey2 struct{}
	)

	t.Run("context key by struct", func(t *testing.T) {
		key1 := TestCtxKey1{}
		ctx := context.WithValue(context.TODO(), key1, "value1")
		t.Log("before, ctx key1 value:", ctx.Value(key1))

		key2 := TestCtxKey2{}
		ctx = context.WithValue(ctx, key2, "value2")
		t.Log("ctx key2 value:", ctx.Value(key2))

		t.Log("after, ctx key1 value:", ctx.Value(key1))
	})

	t.Run("context key by struct, and overwrite", func(t *testing.T) {
		key1 := TestCtxKey1{}
		ctx := context.WithValue(context.TODO(), key1, "value1")
		t.Log("before, ctx key1 value:", ctx.Value(key1))

		key2 := TestCtxKey1{}
		ctx = context.WithValue(ctx, key2, "value2")
		t.Log("ctx key2 value:", ctx.Value(key2))

		t.Log("after, ctx key1 value:", ctx.Value(key1))
	})
}

func TestCtxWithCancelCause(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.TODO())
	go func() {
		ti := time.After(time.Second)
		select {
		case <-ti:
			t.Log("task finish")
		case <-ctx.Done():
			t.Logf("task exit: err=%v, root_cause=%v", ctx.Err(), context.Cause(ctx))
		}
	}()

	time.Sleep(100 * time.Millisecond)
	cancel(fmt.Errorf("custom error"))

	time.Sleep(100 * time.Millisecond)
	t.Log("done")
}

func TestContextAfterFunc(t *testing.T) {
	t.Run("run ctx AfterFunc", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		context.AfterFunc(ctx, func() {
			t.Log("run ctx clearup")
		})

		t.Log("wait...")
		<-ctx.Done()
		time.Sleep(30 * time.Millisecond)
		t.Log("cancelled")
	})

	t.Run("stop ctx AfterFunc", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		stop := context.AfterFunc(ctx, func() {
			t.Log("run ctx clearup")
		})

		select {
		case <-ctx.Done():
			t.Log("cancelled")
		case <-time.After(200 * time.Millisecond):
			if stop() {
				t.Log("stop AfterFunc")
			}
		}

		cancel()
		t.Log("do cancel")
		time.Sleep(30 * time.Millisecond)
		t.Log("finish")
	})
}

// Demo: Unmarshal ptr value

func unMarshalPtr(val []byte, obj any) error {
	// here, obj type is *interface{}
	if string(val) == "null" {
		*obj.(*interface{}) = &TestPerson{}
		return nil
	}

	p := &TestPerson{}
	if err := json.Unmarshal(val, p); err != nil {
		return err
	}

	*obj.(*interface{}) = p
	return nil
}

func TestUnMarshalPtr(t *testing.T) {
	t.Run("null value", func(t *testing.T) {
		var (
			p   *TestPerson
			val = []byte("null")
		)

		getFn := func(obj any) any {
			innerFn := func(innerObj any) {
				err := unMarshalPtr(val, innerObj)
				assert.NoError(t, err)
			}
			innerFn(&obj)
			return obj
		}

		retObj := getFn(p)
		_, ok := retObj.(*TestPerson)
		assert.True(t, ok)
		t.Logf("object: %+v", retObj)
	})

	t.Run("json value", func(t *testing.T) {
		var (
			p   *TestPerson
			val = []byte(`{"name":"foo", "age":33}`)
		)

		getFn := func(obj any) any {
			innerFn := func(innerObj any) {
				err := unMarshalPtr(val, innerObj)
				assert.NoError(t, err)
			}
			innerFn(&obj)
			return obj
		}

		retObj := getFn(p)
		_, ok := retObj.(*TestPerson)
		assert.True(t, ok)
		t.Logf("object: %+v", retObj)
	})
}
