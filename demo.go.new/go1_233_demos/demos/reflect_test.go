package demos

import (
	"reflect"
	"testing"
)

func TestValueCanSet(t *testing.T) {
	x := 10
	v := reflect.ValueOf(x)
	t.Logf("value: can_set=%v", v.CanSet()) // false, 反射得到的是值的副本, 修改副本没有意义

	v = reflect.ValueOf(&x).Elem()
	t.Logf("value ref: can_set=%v", v.CanSet()) // true
	if v.CanSet() {
		v.SetInt(12)
	}
	t.Log("x:", x)
}
