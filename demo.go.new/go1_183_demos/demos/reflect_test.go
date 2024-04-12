package demos_test

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestReflectBasic(t *testing.T) {
	t.Run("get slice type", func(t *testing.T) {
		i := 1
		tof := reflect.TypeOf(i)
		t.Log("type:", tof.String())

		eleTof := reflect.SliceOf(tof)
		t.Log("element type:", eleTof.String())
	})

	t.Run("get pointer type", func(t *testing.T) {
		i := 1
		tof := reflect.TypeOf(i)
		t.Log("type:", tof.String())

		pof := reflect.PtrTo(tof)
		t.Log("type:", pof.Elem().String())
	})
}

func TestStrLenByReflect(t *testing.T) {
	s := "hello world"
	valueOf := reflect.ValueOf(s)
	t.Log("len:", valueOf.Len())

	b := []byte(s)
	valueOf = reflect.ValueOf(b)
	t.Log("len:", valueOf.Len())
}

// Demo: compare by DeepEqual

func TestMapDeepEqual(t *testing.T) {
	t.Run("map same value and struct", func(t *testing.T) {
		m1 := map[string]int{"one": 1, "two": 2}
		m2 := map[string]int{"one": 1, "two": 2}
		result := reflect.DeepEqual(m1, m2)
		t.Log("diff result:", result)
	})

	t.Run("map same value and diff struct", func(t *testing.T) {
		m1 := map[string]int{"one": 1, "two": 2, "three": 3}
		m2 := map[string]int{"two": 2, "three": 3, "one": 1}
		result := reflect.DeepEqual(m1, m2)
		t.Log("diff result:", result)
	})

	t.Run("map diff value", func(t *testing.T) {
		m1 := map[string]int{"one": 1, "two": 2}
		m2 := map[string]int{"two": 2, "one": 3}
		result := reflect.DeepEqual(m1, m2)
		t.Log("diff result:", result)
	})
}

func TestStructDeepEqual(t *testing.T) {
	// TODO:
}

// Demo: new object by reflect

type TestPerson struct {
	Name string `json:"name"`
	Age  uint   `json:"age"`
}

func TestNewObjectByReflect(t *testing.T) {
	b := []byte(`{"name":"foo", "age":31}`)
	pType := getPersonType()

	t.Run("new object by reflect ptr", func(t *testing.T) {
		i := reflect.New(pType).Interface() // ptr, keep type info
		if err := json.Unmarshal(b, i); err != nil {
			t.Fatal(err)
		}

		p := i.(*TestPerson)
		t.Log("struct:", p.Name, p.Age)
	})

	t.Run("new object by reflect struct", func(t *testing.T) {
		i := reflect.New(pType).Elem().Interface() // struct, convert to interface without type info
		if err := json.Unmarshal(b, &i); err != nil {
			t.Fatal(err)
		}

		_, ok := i.(TestPerson)
		t.Log("is struct:", ok)

		m := i.(map[string]any)
		t.Log("map:", m["name"], m["age"])
	})
}

func getPersonType() reflect.Type {
	return reflect.TypeOf(TestPerson{})
}

// Demo: interface{} type check

func TestInterfaceTypeCheck(t *testing.T) {
	type Value struct {
		Id int
	}

	var i interface{} //nolint:gosimple
	i = Value{1}

	tof := reflect.TypeOf(i)
	t.Logf("kind:%s, name:%s", tof.Kind().String(), tof.Name())

	t.Run("switch type check", func(t *testing.T) {
		switch tt := i.(type) {
		case int:
			t.Log("type is int")
		case Value:
			t.Logf("type is Value, id=%d", tt.Id)
		default:
			t.Log("unknown type")
		}
	})

	t.Run("type check by reflect", func(t *testing.T) {
		i = 1
		vof := reflect.ValueOf(i)
		if ok := vof.CanConvert(reflect.TypeOf(0)); ok {
			t.Log("int value:", vof.Int())
		} else {
			t.Log("any value:", vof.Interface())
		}
	})
}

// Demo: check class impl

type MyInterface interface {
	MethodA()
	MethodB()
}

type MyClass struct{}

func NewMyClass() *MyClass {
	return &MyClass{}
}

func (m *MyClass) MethodA() {}

func (m *MyClass) MethodB() {}

func TestInterfaceImpl(t *testing.T) {
	clsType := reflect.TypeOf(NewMyClass())
	interType := reflect.TypeOf((*MyInterface)(nil)).Elem()
	t.Log("interface type:", interType.String())
	t.Log("is impl:", clsType.Implements(interType))
}
