package utils_test

import (
	"testing"

	"demo.apps/utils"
)

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
