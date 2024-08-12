package utils

import (
	"reflect"
	"testing"

	"demo.apps/demos/model"
)

func TestSetUnAddressableField(t *testing.T) {
	var x = 47

	t.Run("can set is True", func(t *testing.T) {
		val := reflect.ValueOf(&x).Elem()
		t.Logf("value=%d, can_set=%v", val.Int(), val.CanSet())
	})

	t.Run("set by unsafe", func(t *testing.T) {
		val := reflect.ValueOf(x)
		t.Logf("value=%d, can_set=%v", val.Int(), val.CanSet())

		SetForflagUnAddrValue(&val)
		t.Logf("after set flag, can_set=%v", val.CanSet())
		val.SetInt(50)
		t.Logf("after change, value=%d", val.Int())
	})
}

func TestSetPrivateFieldOfStruct(t *testing.T) {
	pri := model.NewPriPersonModel("Private", 30)
	pub := model.NewPubPersonModel("Public", 40)

	t.Logf("private struct: %+v", pri)
	t.Logf("public struct: %+v", pub)

	priVal := reflect.ValueOf(&pri).Elem().FieldByName("age")
	pubVal := reflect.ValueOf(&pub).Elem().FieldByName("Age")

	t.Run("print reflect value", func(t *testing.T) {
		t.Logf("private value=%d, can_set=%v", priVal.Int(), priVal.CanSet())
		t.Logf("public value=%d, can_set=%v", pubVal.Int(), pubVal.CanSet())
	})

	t.Run("set private value by unsafe", func(t *testing.T) {
		SetForflagROValue(&priVal)
		priVal.Set(pubVal)
		t.Logf("after set flag: private value=%d, can_set=%v", priVal.Int(), priVal.CanSet())
	})
}
