package demos

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"testing"

	"demo.apps/utils"
)

// demo: json

func TestJsonMarshalForBytes(t *testing.T) {
	type dataHold struct {
		ID string `json:"id"`
		// default marshal slice of bytes as base64
		Bytes []byte `json:"bytes"`
	}

	data := dataHold{
		ID:    "0101",
		Bytes: []byte("hello"),
	}
	b, err := json.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", b)
}

func TestJsonMarshalForRawMessage(t *testing.T) {
	strList := "[1,2,3]"
	maxInt := math.MaxInt64 - 1 // 9223372036854775806
	strObj := fmt.Sprintf(`{"name":"foo","max_int":%d}`, maxInt)

	// #1
	type dataHold1 struct {
		Numbers string
		User    string
	}
	data1 := dataHold1{
		Numbers: strList,
		User:    strObj,
	}
	b, err := json.Marshal(&data1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("json string:", string(b))

	// #2
	type dataHold2 struct {
		Numbers json.RawMessage
		User    json.RawMessage
	}

	data2 := dataHold2{
		Numbers: json.RawMessage(strList),
		User:    json.RawMessage(strObj),
	}
	b, err = json.Marshal(&data2)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("json string with raw message:", string(b))
}

func TestJsonUnmarshalForRawMessage(t *testing.T) {
	maxInt := math.MaxInt64 - 1 // 9223372036854775806
	b := []byte(fmt.Sprintf(`{"name":"foo","max_int":%d}`, maxInt))

	// #1: when unmarshal map number to interface{}, default convert to float64
	m := make(map[string]any)
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	t.Logf("json unmarshal, max int: %v", m["max_int"])

	// #2: json.Number
	m = make(map[string]any)
	if err := utils.JsonLoads(b, &m); err != nil {
		t.Fatal(err)
	}
	num, ok := m["max_int"].(json.Number)
	if !ok {
		t.Fatal("max_int is not json number")
	}
	i, err := num.Int64()
	if err != nil {
		t.Fatal("json number convert to int error")
	}
	t.Logf("json unmarshal with number, max int: %d, %v", i, i == int64(maxInt))

	// #3: json.RawMessage
	type s struct {
		Name string          `json:"name"`
		Num  json.RawMessage `json:"max_int"`
	}
	tmp := s{}
	if err := json.Unmarshal(b, &tmp); err != nil {
		t.Fatal(err)
	}
	t.Log("json unmarshal with raw message, max int:",
		string(tmp.Num), string(tmp.Num) == strconv.Itoa(maxInt))
}

// demo: marshal custom error
//
// refer: http://gregtrowbridge.com/golang-json-serialization-with-interfaces/

type MyError struct {
	err string
}

func (e *MyError) Error() string {
	return e.err
}

func (e *MyError) MarshalJSON() ([]byte, error) {
	m := make(map[string]string)
	m["err_msg"] = e.err
	return json.Marshal(m)
}

func TestJsonMarshalError(t *testing.T) {
	s := struct {
		Id  int   `json:"id"`
		Err error `json:"err"`
	}{
		Id:  1,
		Err: &MyError{err: "mock error"},
	}

	b, err := json.Marshal(&s)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("result:", string(b))
}
