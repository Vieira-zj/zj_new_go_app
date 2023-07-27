package demos

import (
	"encoding/json"
	"fmt"
	"math"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"time"

	"demo.apps/utils"
)

func TestMapCap(t *testing.T) {
	m := make(map[int]string, 2)
	m[1] = "one"
	t.Logf("len=%d", len(m))

	for k, v := range m {
		t.Logf("key=%d, value=%s", k, v)
	}
}

func TestTimeDuration(t *testing.T) {
	duration, err := time.ParseDuration("5m")
	if err != nil {
		t.Fatal(err)
	}
	ti := time.Now().Add(duration)
	t.Log("now after 5m:", ti)
}

// demo: bytes & string

func TestStringMultiReplace(t *testing.T) {
	replace := strings.NewReplacer(" ", "", `\n`, "", `\t`, "")

	str := `{\t"name": "foo",\n\t"age": 31,\n\t"skills:": ["java", "golang"]}`
	result := replace.Replace(str)
	t.Log("result:", result)
}

func TestReuseBytes(t *testing.T) {
	b := []byte("hello")
	t.Log(string(b))

	b = b[:0] // reuse bytes
	t.Log("size:", len(b))
	b = append(b, []byte("foo")...)
	t.Log(string(b))
}

// demo: ref

func TestNilCompare(t *testing.T) {
	var myNil (*byte) = nil

	isNil := true
	if !isNil {
		str := byte(0)
		myNil = &str
	} else {
		t.Log("not init")
	}

	if myNil == nil {
		t.Log("is nil")
	} else {
		t.Log(("is not nil"))
	}
}

func TestRefUpdateSlice(t *testing.T) {
	update := func(fruits []testFruit) {
		for idx := range fruits {
			fruits[idx].name += "-test"
		}
	}

	fruits := []testFruit{
		{"apple"},
		{"pair"},
	}
	t.Log("fruits:", fruits)
	update(fruits)
	t.Log("new fruits:", fruits)
}

func TestRefUpdateMap(t *testing.T) {
	update := func(fruits map[string]testFruit) {
		for key, fruit := range fruits {
			fruit.name += "-test"
			fruits[key] = fruit
		}
	}

	fruits := map[string]testFruit{
		"apple": {"apple"},
		"pair":  {"pair"},
	}
	t.Log("fruits:", fruits)
	update(fruits)
	t.Log("new fruits:", fruits)
}

// demo: struct

type testFruit struct {
	name string
}

type testStudent struct {
	name   string
	age    int
	tags   []string
	bag    []testFruit
	scores map[string]int
}

func NewTestStudent(name string, age int) testStudent {
	return testStudent{
		name:   name,
		age:    age,
		tags:   make([]string, 0, 2),
		bag:    []testFruit{{"apple"}, {"banana"}},
		scores: make(map[string]int, 2),
	}
}

func (s *testStudent) setName(name string) {
	s.name = name
}

func (s *testStudent) setAge(age int) {
	s.age = age
}

func (s *testStudent) addTag(tag string) {
	s.tags = append(s.tags, tag)
}

func (s testStudent) updateBag(idx int, name string) {
	s.bag[idx].name = name
}

func (s testStudent) addScore(key string, value int) {
	s.scores[key] = value
}

func (s testStudent) String() string {
	fruits := ""
	for idx, f := range s.bag {
		if idx == 0 {
			fruits = f.name
			continue
		}
		fruits += "," + f.name
	}
	return fmt.Sprintf("name=%s,age=%d,tags=%v,fruit=[%s],scores=%+v", s.name, s.age, s.tags, fruits, s.scores)
}

func updateStudent(s testStudent) testStudent {
	s.setName("bar")
	s.setAge(19)
	s.addTag("p2")
	s.updateBag(0, "pair")
	s.addScore("cn", 98)
	return s
}

func updateStudentRef(s *testStudent) {
	s.setName("bar")
	s.setAge(19)
	s.addTag("p2")
	s.updateBag(1, "cherry")
	s.addScore("cn", 98)
}

func TestUpdateStruct(t *testing.T) {
	s1 := NewTestStudent("foo", 13)
	s1.addTag("p1")
	s1.addScore("en", 91)
	t.Logf("s1 [%p]: %s", &s1, s1.String())

	s2 := updateStudent(s1)
	t.Log("after update")
	t.Log("s1:", s1.String())
	t.Log("s2:", s2.String())
}

func TestUpdateStructRef(t *testing.T) {
	s := NewTestStudent("foo", 13)
	s.addTag("p1")
	s.addScore("en", 91)
	t.Logf("s [%p]: %s", &s, s.String())

	updateStudentRef(&s)
	t.Log("after update")
	t.Logf("s [%p]: %s", &s, s.String())
}

// demo: copy of struct

func TestCopyOfStruct(t *testing.T) {
	students := make([]testStudent, 0, 2)
	for i := 0; i < 2; i++ {
		students = append(students, testStudent{
			name: fmt.Sprintf("tester_%d", i),
			age:  30 + i,
			tags: []string{strconv.Itoa(i)},
		})
	}

	copied := make([]testStudent, 2)
	copy(copied, students)

	for i := range copied {
		copied[i].name = fmt.Sprintf("tester_%d_copied", i)
		copied[i].age += 5
		copied[i].tags[0] = fmt.Sprintf("%d_copied", i)
	}

	t.Log("src students:")
	for _, s := range students {
		t.Log(s.name, s.age, s.tags)
	}
	t.Log("dest students:")
	for _, s := range copied {
		t.Log(s.name, s.age, s.tags)
	}
}

// demo: json

func TestJsonMarshalForRawMessage(t *testing.T) {
	strList := "[1,2,3]"
	maxInt := math.MaxInt64 - 1 // 9223372036854775806
	strObj := fmt.Sprintf(`{"name":"foo","max_int":%d}`, maxInt)

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

	// when unmarshal map number to interface{}, default convert to float64
	m := make(map[string]any)
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	t.Logf("json unmarshal, max int: %v", m["max_int"])

	// json.Number
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

	// json.RawMessage
	type s struct {
		Name string          `json:"name"`
		Num  json.RawMessage `json:"max_int"`
	}
	tmp := s{}
	if err := json.Unmarshal(b, &tmp); err != nil {
		t.Fatal(err)
	}
	t.Log("json unmarshal with raw message, max int:", string(tmp.Num), string(tmp.Num) == strconv.Itoa(maxInt))
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

// demo: goroutine

func TestRecoverFromPanic(t *testing.T) {
	ch := make(chan struct{})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// print stack from recover
				fmt.Println("recover err:", r)
				fmt.Printf("stack:\n%s", debug.Stack())
			}
			ch <- struct{}{}
		}()
		fmt.Println("goroutine run...")
		time.Sleep(time.Second)
		panic("mock err")
	}()

	fmt.Println("wait...")
	<-ch
	fmt.Println("recover demo done")
}
