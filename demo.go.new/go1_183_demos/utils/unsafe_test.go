package utils

import (
	"reflect"
	"testing"
	"unsafe"

	"demo.apps/demos/model"
)

func TestGetSliceElement(t *testing.T) {
	// 1. go 中切片其实是一个 runtime 中的 slice 结构体
	// 2. append 操作 slice 底层的数组后, 改变了数组的长度或者容量, 所以需要重新赋值给原先的切片. 如果切片发生扩容, 原先的数组地址也要发生变化

	s := make([]int64, 1, 2)
	_ = append(s, 10)
	t.Log("slice:", s)

	// 强制访问 s 的第 1 个元素
	baseAddr := unsafe.Pointer(&s[0])
	offset := unsafe.Sizeof(s[0])
	ele := *((*int64)(unsafe.Pointer(uintptr(baseAddr) + offset)))
	t.Log("1st element:", ele)
}

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