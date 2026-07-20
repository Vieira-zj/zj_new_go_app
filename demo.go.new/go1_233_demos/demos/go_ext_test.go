package demos

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Demo: Json

func TestJsonUnmarshalPartly(t *testing.T) {
	type Addr struct {
		Country string `json:"country"`
		City    string `json:"city"`
	}
	type Person struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Addr Addr   `json:"addr"`
	}

	p := Person{ID: 101, Name: "Foo"}
	addr := `{"addr":{"country":"cn","city":"wuhan"}}`
	err := json.Unmarshal([]byte(addr), &p)
	assert.NoError(t, err)
	t.Logf("person: %+v", p)
}

func TestJsonMarshalTags(t *testing.T) {
	type Person struct {
		ID    int    `json:"id,string"`
		Name  string `json:"name"`
		Level int    `json:"level,omitzero"`
		Desc  string `json:"description,omitempty"`
		// tag:omitzero checks for time.Time IsZero()
		UpdatedBy time.Time `json:"update_by,omitzero"`
	}

	t.Run("marshal with tag fields", func(t *testing.T) {
		p := Person{
			ID:        102,
			Name:      "Foo",
			Level:     31,
			Desc:      "A person description",
			UpdatedBy: time.Now(),
		}
		b, err := json.Marshal(&p)
		assert.NoError(t, err)
		t.Log("json:", string(b))
	})

	t.Run("marshal without tag fields", func(t *testing.T) {
		p := Person{
			ID:   102,
			Name: "Foo",
		}
		b, err := json.Marshal(&p)
		assert.NoError(t, err)
		t.Log("json:", string(b))
	})
}

func TestJsonOmitTag(t *testing.T) {
	// 精确控制零值用 omitzero, 常规空值忽略用 omitempty
	// 通过 IsZero() 自定义零值判断
	type Data struct {
		Field1 string    `json:"field1,omitempty"` // omit
		Field2 string    `json:"field2,omitzero"`  // omit
		Time1  time.Time `json:"time1,omitempty"`  // "time1": "0001-01-01T00:00:00Z"
		Time2  time.Time `json:"time2,omitzero"`   // omit
		Slice1 []int     `json:"slice1,omitempty"` // omit
		Slice2 []int     `json:"slice2,omitzero"`  // "slice2": []
	}

	data := Data{
		Field1: "",
		Field2: "",
		Time1:  time.Time{},
		Time2:  time.Time{},
		Slice1: []int{},
		Slice2: []int{},
	}

	b, err := json.MarshalIndent(data, "", "  ")
	assert.NoError(t, err)
	t.Logf("marshal results:\n%s", b)
}

func TestJsonStream01(t *testing.T) {
	type User struct {
		Name  string
		Score int
	}
	users := []User{
		{"Carol", 90},
		{"Alice", 95},
		{"Bob", 80},
	}
	b, err := json.Marshal(&users)
	require.NoError(t, err, "prepare json bytes failed")

	dec := json.NewDecoder(bytes.NewReader(b))
	token, err := dec.Token()
	assert.NoError(t, err)
	assert.True(t, json.Delim('[') == token, "expected array start delimiter")

	for dec.More() {
		var u User
		err = dec.Decode(&u)
		assert.NoError(t, err)
		t.Log("decode user:", u)
	}

	token, err = dec.Token()
	assert.NoError(t, err)
	assert.True(t, json.Delim(']') == token, "expected array end delimiter")
}

func TestJsonStream02(t *testing.T) {
	jsonStreamReader := func() io.Reader {
		pr, pw := io.Pipe()
		// pipe 生产者和消费者必须身处不同的 goroutine
		go func() {
			defer pr.Close()

			data := map[string]string{
				"status":  "ok",
				"message": "processing large stream ...",
			}
			if err := json.NewEncoder(pw).Encode(data); err != nil {
				pw.CloseWithError(err)
			}
		}()
		return pr
	}

	r := jsonStreamReader()
	b, err := io.ReadAll(r)
	assert.ErrorIs(t, err, io.ErrClosedPipe)
	t.Logf("result: %s", b)
}

// Demo: Reflect

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

// Demo: Reg Exp

func TestRegExpMatch(t *testing.T) {
	var emailRegex = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+.[a-z]{2,4}$`)

	t.Run("validate email", func(t *testing.T) {
		ok := emailRegex.MatchString("xxxx@google.com")
		t.Log("is matched:", ok)

		ok = emailRegex.MatchString("google.com")
		t.Log("is matched:", ok)
	})
}

func TestRegExpFind(t *testing.T) {
	var idRegex = regexp.MustCompile(`ID:(\d+)`)

	t.Run("find in long content", func(t *testing.T) {
		longContent := "IDs,ID:001,ID:002,ID:003,ID:004,ID:005,ID:006"
		matches := idRegex.FindStringSubmatch(longContent)
		// 这里 id 引用整个 longContent
		// id := matches[1]

		id := strings.Clone(matches[1])
		t.Log("1st matched id:", id)
	})
}
