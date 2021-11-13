package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type reflectUser struct {
	Name   string
	Age    int `json:"age,omitempty"`
	Meta   reflectMeta
	Skills []string
}

type reflectMeta struct {
	ID   int
	Desc string
}

func TestJsonDumpWithDefaultValue(t *testing.T) {
	data := reflectUser{}
	b, err := json.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}

func TestReflectSliceDemo(t *testing.T) {
	sl := []string{"a", "b"}
	typeOf := reflect.TypeOf(sl)
	fmt.Println("slice type:", typeOf.Kind().String())
	elem := typeOf.Elem()
	fmt.Println("slice element type:", elem.Kind().String())

	valueOf := reflect.ValueOf(sl)
	items := make([]interface{}, 0, valueOf.Len())
	for i := 0; i < valueOf.Len(); i++ {
		item := valueOf.Index(i)
		fmt.Println(item.Kind())
		items = append(items, item.Interface())
	}
	fmt.Println("slice values:", items)
}

func TestReflectStructDemo(t *testing.T) {
	user := &reflectUser{
		Name: "foo",
		Age:  29,
	}

	valueOf := reflect.ValueOf(user)
	if valueOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
	}

	for i := 0; i < valueOf.NumField(); i++ {
		// valueOf => 字段的字段的类型和interface值
		fValue := valueOf.Field(i)
		// typeOf => 字段的名称和tag定义
		fType := valueOf.Type().Field(i)
		fmt.Printf("%s[%s]: %v\n", fType.Name, fValue.Kind(), fValue.Interface())
	}
}

func TestReflectStructSetDemo(t *testing.T) {
	// Meta fields must be "public"
	m := &reflectMeta{
		ID:   1,
		Desc: "init",
	}

	valueOf := reflect.ValueOf(m)
	if valueOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
	}
	fmt.Println(valueOf.Kind().String())

	for i := 0; i < valueOf.NumField(); i++ {
		field := valueOf.Field(i)
		switch field.Kind() {
		case reflect.Int:
			field.SetInt(2)
		case reflect.String:
			field.SetString("new")
		default:
			t.Fatal("invalid field type")
		}
	}
	fmt.Printf("new: %+v\n", valueOf.Interface())
}

//
// Struct to Map
//

func TestReflectStructToMapV1(t *testing.T) {
	user := reflectUser{
		Name: "foo",
		Age:  30,
		Meta: reflectMeta{
			ID:   1,
			Desc: "reflect_struct_test",
		},
		Skills: []string{"python", "golang"},
	}

	t.Run("#1. reflect struct", func(t *testing.T) {
		res, err := ReflectStructToMapV1(user)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("reflect results: %+v\n", res)
	})

	t.Run("#2. reflect ptr", func(t *testing.T) {
		res, err := ReflectStructToMapV1(&user)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("reflect results: %+v\n", res)
	})
}

type userInfo struct {
	Name    string  `json:"name"`
	Age     int     `json:"age"`
	Profile profile `json:"profile"`
}

type profile struct {
	Hobby string `json:"hobby" structs:"hobby"`
}

func TestReflectStructToMapV2(t *testing.T) {
	user := userInfo{Name: "foo", Age: 18, Profile: profile{"bar"}}
	resMap, err := ReflectStructToMapV2(user, "json")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("reflect struct: %+v\n", resMap)
}

//
// Map to Struct
//

type fruit struct {
	ID    int     `key:"id"`
	Name  string  `key:"name"`
	Price float32 `key:"price"`
}

func TestReflectMapToStruct(t *testing.T) {
	data := map[string]interface{}{
		"id":    1001,
		"name":  "apple",
		"price": 16, // int
		"meta":  "desc",
	}

	fruit := &fruit{}
	if err := ReflectMapToStruct(data, fruit); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", *fruit)
}

//
// Others
//

func TestReflectSliceToString(t *testing.T) {
	t.Run("#1. reflect string", func(t *testing.T) {
		sl := []string{"one", "two", "three"}
		res, err := ReflectSliceToString(sl)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("slice results:", res)
	})

	t.Run("#2. reflect struct", func(t *testing.T) {
		user := reflectUser{
			Name: "foo",
			Age:  30,
			Meta: reflectMeta{
				ID:   1,
				Desc: "reflect_slice_test",
			},
			Skills: []string{"python", "golang"},
		}
		sl := []reflectUser{user}
		res, err := ReflectSliceToString(sl)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("slice results:", res)
	})
}

func TestStringToSliceByte(t *testing.T) {
	out := StringToSliceByte("hello world")
	for i := 0; i < len(out); i++ {
		fmt.Printf("%c", out[i])
	}
	fmt.Println()
}
