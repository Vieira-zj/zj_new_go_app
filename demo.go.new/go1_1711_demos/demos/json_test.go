package demos

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/imdario/mergo"
	"github.com/stretchr/testify/assert"
)

func TestMarshalFunc(t *testing.T) {
	// json.Marshal unsupported type: func()
	type caller struct {
		Name string `json:"name"`
		Fn   func() `json:"func"`
	}

	c := &caller{
		Name: "helloworld",
		Fn: func() {
			fmt.Println("helloworld")
		},
	}
	b, err := json.Marshal(c)
	assert.NoError(t, err)
	fmt.Printf("caller: %s\n", b)
}

func TestMarshalStruct(t *testing.T) {
	type Base struct {
		Id string `json:"id"`
	}
	type Student struct {
		Base
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	s := Student{
		Base: Base{
			Id: "001",
		},
		Name: "foo",
		Age:  31,
	}
	s.Id = "002"
	b, err := json.MarshalIndent(s, "", "  ")
	assert.NoError(t, err)
	t.Log("student:\n", string(b))
}

func TestStructToMap01(t *testing.T) {
	type person struct {
		ID   uint8  `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	p := person{
		ID:   1,
		Name: "foo",
		Age:  31,
	}
	b, err := json.Marshal(&p)
	if err != nil {
		t.Fatal(err)
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	fmt.Println("src map:", m)

	// add key, value
	m["comment"] = "for test"
	// update key name
	m["identity"] = m["id"]
	delete(m, "id")

	b, err = json.MarshalIndent(&m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("dst map:\n%s\n", b)
}

func TestStructToMap02(t *testing.T) {
	type Fruit struct {
		Name     string `json:"name"`
		PriceTag string `json:"priceTag"`
	}
	type Basket struct {
		Fruits map[string]Fruit `json:"fruits"`
	}

	jsonData := `{
		"fruits": {
			"1": {
				"name": "Apple",
				"priceTag": "$1"
			},
			"2": {
				"name": "Pear",
				"priceTag": "$1.5"
			}
		}
	}`
	var basket Basket
	err := json.Unmarshal([]byte(jsonData), &basket)
	assert.NoError(t, err)
	for _, fruit := range basket.Fruits {
		fmt.Println(fruit.Name+":", fruit.PriceTag)
	}
}

func TestStructToMap03(t *testing.T) {
	jsonData := `{"name":"Foo", "age":6, "parents":["Alice","Bob"]}`
	var val interface{}
	err := json.Unmarshal([]byte(jsonData), &val)
	assert.NoError(t, err)

	data := val.(map[string]interface{})
	for k, v := range data {
		switch v := v.(type) {
		case string:
			fmt.Println(k, v, "(string)")
		case float64:
			fmt.Println(k, v, "(float64)")
		case []interface{}:
			fmt.Println(k, "(slice):")
			for i, u := range v {
				fmt.Println("\t", i, u)
			}
		default:
			fmt.Println(k, v, "unknown")
		}
	}
}

func TestJsonUnmarshalAsInterface(t *testing.T) {
	text := `{"env":["test", "uat"]}`
	res := make(map[string]interface{})
	err := json.Unmarshal([]byte(text), &res)
	assert.NoError(t, err)

	if env, ok := res["env"]; ok {
		valueOf := reflect.ValueOf(env)
		fmt.Println("value type:", valueOf.Type().Kind())
		itemValueOf := valueOf.Index(0)
		fmt.Printf("slice item value type: %s\n", itemValueOf.Type().Kind())
		if values, ok := env.([]interface{}); ok && len(values) > 0 {
			t.Log("1st env:", values[0])
		}
	} else {
		t.Fatal("no env found")
	}
}

func TestJsonDecode(t *testing.T) {
	type Message struct {
		Name, Text string
	}

	const stream = `
	{"name": "Ed", "text": "Knock knock."}
	{"name": "Sam", "text": "Who's there?"}
	{"name": "Ed", "text": "Go fmt."}
	{"name": "Sam", "text": "Go fmt who?"}
	{"name": "Ed", "text": "Go fmt yourself!"}
	`

	dec := json.NewDecoder(strings.NewReader(stream))
	t.Log("output:")
	for {
		var m Message
		if err := dec.Decode(&m); err != nil {
			if errors.Is(err, io.EOF) {
				t.Log("end of stream")
				break
			}
			t.Fatal(err)
		}
		t.Log(m.Name, m.Text)
	}
}

//
// 1. mergo 不会复制非导出字段
// 2. map 使用时候，对应的key字段默认是小写的
// 3. mergo 可以嵌套赋值
//

type Student struct {
	Name string
	Num  int
	Age  int
}

func TestMergoStructToMap(t *testing.T) {
	student := Student{
		Name: "foo",
		Num:  1,
		Age:  18,
	}

	m := make(map[string]interface{})
	if err := mergo.Map(&m, student); err != nil {
		t.Fatal(err)
	}
	t.Logf("map m: %+v", m)
}

func TestMergoMapToStruct(t *testing.T) {
	m := map[string]interface{}{
		"name": "bar",
		"num":  2,
		"age":  20,
	}

	s := Student{}
	if err := mergo.Map(&s, m); err != nil {
		t.Fatal(err)
	}
	t.Logf("struct student: %+v", s)
}
