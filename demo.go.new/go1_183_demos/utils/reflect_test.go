package utils_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"demo.apps/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetPkgName(t *testing.T) {
	type Empty struct{}

	tp := reflect.TypeOf(Empty{})
	t.Log("package path:", tp.PkgPath())
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

func TestWrapFunc(t *testing.T) {
	helloFn := func(name string) string {
		fmt.Println("helloFn run")
		return "Hello, " + name
	}

	addFn := func(i, j int) string {
		fmt.Println("addFn run")
		return strconv.Itoa(i + j)
	}

	t.Run("wrap invalid fn", func(t *testing.T) {
		_, err := utils.WrapFunc("1")
		assert.Error(t, err, "input arg is not a func")
	})

	t.Run("warp fn", func(t *testing.T) {
		fn, err := utils.WrapFunc(helloFn)
		assert.NoError(t, err)
		if hello, ok := fn.(func(name string) string); ok {
			t.Log(hello("foo"))
		}

		fn, err = utils.WrapFunc(addFn)
		assert.NoError(t, err)
		if add, ok := fn.(func(i, j int) string); ok {
			t.Log("add result:", add(2, 4))
		}
	})
}
