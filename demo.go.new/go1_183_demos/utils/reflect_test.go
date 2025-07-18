package utils_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"demo.apps/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetPkgName(t *testing.T) {
	type Empty struct{}

	tp := reflect.TypeOf(Empty{})
	t.Log("package path:", tp.PkgPath())
}

func TestIsNil01(t *testing.T) {
	var x any
	var y *int = nil
	x = y

	typOf := reflect.TypeOf(x)
	t.Logf("kind:%v, type:%s", typOf.Kind(), typOf.String())

	t.Run("nil any var check", func(t *testing.T) {
		runCase := func(a any) {
			// 即使接口持有的值为 nil, 也并不意味着接口本身为 nil
			t.Log("x:", a)
			t.Log("x != nil", a != nil) // true
		}
		runCase(x)
	})

	t.Run("is nil check", func(t *testing.T) {
		t.Log("IsNil:", utils.IsNil(x))
	})
}

type MyInterface interface {
	apply()
}

type MyInterfaceImpl struct{}

func (*MyInterfaceImpl) apply() {}

// 在编译阶段检查 impl 接口实现
var _ MyInterface = (*MyInterfaceImpl)(nil)

func TestIsNil02(t *testing.T) {
	var x MyInterface
	var y *MyInterfaceImpl = nil
	x = y

	typOf := reflect.TypeOf(x)
	t.Logf("kind:%v, type:%s", typOf.Kind(), typOf.String())

	t.Run("nil interface var check", func(t *testing.T) {
		runCase := func(a MyInterface) {
			t.Log("x:", a)
			t.Log("x != nil", a != nil) // true
		}
		runCase(x)
	})

	t.Run("is nil check", func(t *testing.T) {
		t.Log("IsNil:", utils.IsNil(x))
	})
}

type MyPersonImpl struct {
	Name string
	Age  int
}

func TestIsEmptyStruct(t *testing.T) {
	t.Run("empty person", func(t *testing.T) {
		p := MyPersonImpl{}
		assert.True(t, utils.IsEmptyStruct(p))
	})

	t.Run("empty person pointer", func(t *testing.T) {
		p := MyPersonImpl{}
		assert.True(t, utils.IsEmptyStruct(&p))
	})

	t.Run("person with name", func(t *testing.T) {
		p := MyPersonImpl{}
		p.Name = "foo"
		assert.False(t, utils.IsEmptyStruct(p))
	})

	t.Run("person with age", func(t *testing.T) {
		p := MyPersonImpl{}
		p.Age = 40
		assert.False(t, utils.IsEmptyStruct(p))
	})
}

func TestTrimStringFields(t *testing.T) {
	type TestStruct struct {
		StrValue1 string
		StrValue2 string
		strValue3 string
		IntValue  int
	}

	s := TestStruct{
		StrValue1: "value1 ",
		StrValue2: "  value2 ",
		strValue3: " value3",
		IntValue:  3,
	}

	if err := utils.TrimStringFields(s); err != nil {
		t.Log("expect error:", err)
	}

	t.Logf("before trim: %+v", s)
	if err := utils.TrimStringFields(&s); err != nil {
		t.Fatal(err)
	}
	t.Logf("after trim: %+v", s)
}

func TestSlicesContains(t *testing.T) {
	t.Run("same size slice", func(t *testing.T) {
		s1 := []string{"a", "b", "c"}
		s2 := []string{"a", "c", "b"}
		result := utils.SlicesContains[string](s1, s2)
		t.Log("result:", result)
	})

	t.Run("diff size slice", func(t *testing.T) {
		s1 := []string{"a", "b", "c"}
		s2 := []string{"a", "y", "c", "b", "x"}
		result := utils.SlicesContains(s1, s2)
		t.Log("result:", result)
	})

	t.Run("diff slice", func(t *testing.T) {
		s1 := []string{"a", "b1", "c"}
		s2 := []string{"a", "c", "b"}
		result := utils.SlicesContains(s1, s2)
		t.Log("result:", result)
	})
}

func TestGetFnDeclaration(t *testing.T) {
	for _, fn := range []any{
		utils.GetLocalIPAddr,
		utils.GetCallerInfo,
	} {
		result, err := utils.GetFnDeclaration(fn)
		if err != nil {
			t.Fatal(err)
		}

		b, _ := json.Marshal(&result)
		t.Logf("func desc: %s", b)
	}
}

func TestReflectWrapFunc(t *testing.T) {
	helloFn := func(name string) string {
		fmt.Println("helloFn run")
		time.Sleep(200 * time.Millisecond)
		return "Hello, " + name
	}

	addFn := func(i, j int) string {
		fmt.Println("addFn run")
		time.Sleep(250 * time.Millisecond)
		return strconv.Itoa(i + j)
	}

	t.Run("wrap invalid func", func(t *testing.T) {
		_, err := utils.ReflectWrapFunc("1")
		assert.Error(t, err, "input arg is not a func")
	})

	t.Run("wrap func", func(t *testing.T) {
		fn, err := utils.ReflectWrapFunc(helloFn)
		assert.NoError(t, err)
		if hello, ok := fn.(func(name string) string); ok {
			t.Log(hello("foo"))
		}

		fn, err = utils.ReflectWrapFunc(addFn)
		assert.NoError(t, err)
		if add, ok := fn.(func(i, j int) string); ok {
			t.Log("add result:", add(2, 4))
		}
	})
}

func TestReflectDeepCopy(t *testing.T) {
	data := map[string]struct {
		Id  string
		Age int
	}{
		"foo": {"1010", 31},
		"bar": {"1011", 29},
	}

	dst := utils.ReflectDeepCopy(data)
	dstMap, ok := dst.(map[string]struct {
		Id  string
		Age int
	})
	assert.True(t, ok)

	for key, val := range dstMap {
		b, err := json.Marshal(val)
		assert.NoError(t, err)
		t.Logf("name:%s, %s", key, b)
	}
}
