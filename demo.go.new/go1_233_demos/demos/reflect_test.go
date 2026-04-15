package demos

import (
	"reflect"
	"testing"
)

func TestSetIntValue(t *testing.T) {
	x := 10
	v := reflect.ValueOf(x)
	t.Logf("value: can_set=%v", v.CanSet()) // false (反射得到的是值的副本, 修改副本没有意义)

	v = reflect.ValueOf(&x).Elem()
	t.Logf("value ref: can_set=%v", v.CanSet()) // true
	if v.CanSet() {
		v.SetInt(12)
	}
	t.Log("x after updated:", x)
}

func TestSetStructField(t *testing.T) {
	type person struct {
		name string
		Age  int
	}

	p := person{name: "Bar", Age: 41}
	val := reflect.ValueOf(&p) // use ref for updating

	nameField := val.Elem().FieldByName("name")
	if nameField.IsValid() {
		t.Log("name before update:", nameField.String())
		if nameField.CanSet() {
			t.Log("update name")
			nameField.SetString("Foo")
		}
	}

	ageField := val.Elem().FieldByName("Age")
	if ageField.IsValid() {
		t.Log("age before update:", ageField.Int())
		if ageField.CanSet() {
			t.Log("update age")
			ageField.SetInt(35)
		}
	}

	t.Log("p after updated:", p)
}
