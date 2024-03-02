package demos_test

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
		// marshal bytes as base64
		Bytes []byte `json:"bytes"`
		// marshal as raw struct
		RawBytes  json.RawMessage `json:"raw_bytes"`
		RawString json.RawMessage `json:"raw_string"`
	}

	data := dataHold{
		ID:        "0101",
		Bytes:     []byte("hello"),
		RawBytes:  []byte(`{"name":"foo"}`),
		RawString: json.RawMessage(`{"id":"0101"}`),
	}
	b, err := json.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s", b)
}

func TestJsonMarshalForRawMsg(t *testing.T) {
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

func TestJsonUnmarshalForRawMsg(t *testing.T) {
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

// demo: custom json parse

type OrderId uint64

func (o OrderId) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprint(o)), nil
}

func (o *OrderId) UnmarshalText(text []byte) error {
	result, err := strconv.ParseUint(string(text), 10, 64)
	if err != nil {
		return err
	}
	*o = OrderId(result)
	return nil
}

type TestJsonData struct {
	ID      uint64  `json:"id"`
	OrderID OrderId `json:"order_id"`
}

func TestCustomJsonParse(t *testing.T) {
	data := TestJsonData{
		ID:      1010,
		OrderID: 10101,
	}
	b, err := json.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("results:", string(b))

	d := TestJsonData{}
	err = json.Unmarshal(b, &d)
	if err != nil {
		t.Log("error:", err)
	}
	t.Log(d.ID, d.OrderID)
}

// demo: marshal error by custom json parse
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
