package utils_test

import (
	"reflect"
	"testing"

	"demo.apps/utils"
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
