package demos_test

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestClearupCase(t *testing.T) {
	t.Cleanup(func() {
		t.Log("case clear")
	})

	if ok := false; !ok {
		t.Fatal("mock fatal")
	}
	t.Log("case run")
}

func TestSwitchConds(t *testing.T) {
	getNumberDesc := func(num int) string {
		switch num {
		case 1, 2, 3:
			return "num <= 3"
		case 4, 5:
			return "3 < num <= 5"
		default:
			return "num > 5"
		}
	}

	for i := 2; i < 7; i++ {
		desc := getNumberDesc(i)
		t.Logf("%d: %s", i, desc)
	}
}

func TestTypeAssert(t *testing.T) {
	var s, m any

	t.Run("slice", func(t *testing.T) {
		s = []int{1}
		switch s.(type) {
		case []any:
			t.Log("type: []any")
		case []int:
			t.Log("type: []int")
		default:
			t.Log("unknown type")
		}
	})

	t.Run("map", func(t *testing.T) {
		m = map[string]int{"one": 1}
		switch m.(type) {
		case map[string]any:
			t.Log("type: map[string]any")
		default:
			t.Log("unknown type")
		}
	})
}

func TestTimeDuration(t *testing.T) {
	duration, err := time.ParseDuration("5m")
	if err != nil {
		t.Fatal(err)
	}
	ti := time.Now().Add(duration)
	t.Log("now after 5m:", ti)
}

func TestGoSlog(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	logger.Debug("text debug level log", "uid", 1002)
	logger.Info("text info level log", "uid", 1002)

	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	jsonLogger.Debug("json info level log", "uid", 1002)
	jsonLogger.Info("json info level log", "uid", 1002)
}

// Demo: defer

func TestDeferFn01(t *testing.T) {
	testFn := func() func() {
		t.Log("test fn")
		return func() {
			t.Log("wrapped test fn")
		}
	}

	defer testFn()()
	t.Log("start test defer fn")
	time.Sleep(200 * time.Millisecond)
	t.Log("end test defer fn")
}

type myTestStruct struct {
	t *testing.T
}

func (s *myTestStruct) fn1() *myTestStruct {
	s.t.Log("fn1 invoke")
	return s
}

func (s *myTestStruct) fn2() *myTestStruct {
	s.t.Log("fn2 invoke")
	return s
}

func TestDeferFn02(t *testing.T) {
	s := &myTestStruct{t}
	defer s.fn1().fn2()

	t.Log("start test defer struct fn")
	time.Sleep(200 * time.Millisecond)
	t.Log("end test defer struct fn")
}

// Demo: slice

func TestSliceInitByIndex(t *testing.T) {
	s := []string{
		2: "two", // index:value
		3: "three",
		1: "one",
		0: "zero",
	}
	t.Logf("len=%d, cap=%d", len(s), cap(s))
	for idx, val := range s {
		t.Logf("%d:%s", idx, val)
	}

	t.Run("case1", func(t *testing.T) {
		s1 := s[:0]
		t.Logf("len=%d, cap=%d: %v", len(s1), cap(s1), s1)
	})

	t.Run("case2", func(t *testing.T) {
		s1 := s[:2]
		t.Logf("len=%d, cap=%d: %v", len(s1), cap(s1), s1)
	})

	t.Run("case3", func(t *testing.T) {
		s1 := s[2:]
		t.Logf("len=%d, cap=%d: %v", len(s1), cap(s1), s1)
	})

	t.Run("case4", func(t *testing.T) {
		// cap only relate to start index
		s1 := s[1:3]
		t.Logf("len=%d, cap=%d: %v", len(s1), cap(s1), s1)
	})
}

func TestSliceAppend(t *testing.T) {
	t.Run("append", func(t *testing.T) {
		s := make([]int, 1, 2)
		s[0] = -1
		s2 := append(s, 2)
		s2[0] = -2

		// replace
		s3 := append(s, 3)
		s3[0] = -3
		// s,s1,s3 have same address of array
		for i, sl := range [][]int{s, s2, s3} {
			t.Logf("s%d (%p): %v", i+1, sl, sl)
		}
	})

	t.Run("append with scale", func(t *testing.T) {
		s := make([]int, 1, 2)
		s[0] = -1
		s2 := append(s, 2)
		s2[0] = -2

		// scale, and return new address of slice
		s3 := append(s2, 3)
		s3[0] = -3
		for i, sl := range [][]int{s, s2, s3} {
			t.Logf("s%d (%p): %v", i+1, sl, sl)
		}
	})
}

func TestSliceAddValue(t *testing.T) {
	size := 3
	s := make([]int, size, size*2)
	for i := 0; i < size; i++ {
		s[i] = i
	}

	s = s[0 : size+1]
	s[size] = 4
	t.Logf("len=%d, cap=%d, values: %v", len(s), cap(s), s)

	s = append(s, 5)
	t.Logf("len=%d, cap=%d, values: %v", len(s), cap(s), s)
}

func TestSliceCopy(t *testing.T) {
	type T struct {
		id    int
		value string
	}

	const size = 3

	t.Run("slice of values", func(t *testing.T) {
		s := make([]T, size)
		for i := 0; i < size; i++ {
			s[i] = T{id: i, value: strconv.Itoa(i)}
		}
		dst := make([]T, size)
		_ = copy(dst, s)

		s[0].value = "zero"
		s[1].value = "one"

		t.Log("unchange:")
		for _, item := range dst {
			t.Log(item.id, item.value)
		}
	})

	t.Run("slice of refs", func(t *testing.T) {
		s := make([]*T, size)
		for i := 0; i < size; i++ {
			s[i] = &T{id: i, value: strconv.Itoa(i)}
		}
		dst := make([]*T, size)
		_ = copy(dst, s)

		s[0].value = "zero"
		s[1].value = "one"

		t.Log("changed:")
		for _, item := range dst {
			t.Log(item.id, item.value)
		}
	})
}

// Demo: map

func TestMapCap(t *testing.T) {
	m := make(map[int]string, 2)
	m[1] = "one"
	t.Logf("len=%d", len(m))

	for k, v := range m {
		t.Logf("key=%d, value=%s", k, v)
	}
}

func TestMapPtrAsKey(t *testing.T) {
	type num struct {
		id    int
		value string
	}

	one := &num{id: 1, value: "one"}
	two := &num{id: 2, value: "two"}

	// use address as map key
	m := map[*num]string{
		one: "num_one",
		two: "num_two",
	}

	t.Run("iterator", func(t *testing.T) {
		for k, v := range m {
			t.Log(k.id, k.value, v)
		}
	})

	t.Run("get exist", func(t *testing.T) {
		if v, ok := m[one]; ok {
			t.Log(v)
		} else {
			t.Fatal("not found")
		}
	})

	t.Run("get new", func(t *testing.T) {
		one = &num{id: 1, value: "one"}
		if v, ok := m[one]; ok {
			t.Log(v)
		} else {
			t.Log("not found")
		}
	})
}

// Demo: bits, number

func TestHexToDecimal(t *testing.T) {
	val := 0xff
	result := strconv.FormatInt(int64(val), 10)
	t.Logf("decimal result: %d, %s", int64(val), result)
}

func TestCalBits(t *testing.T) {
	t.Log(1 << 0)
	t.Log(1 << 1)
	t.Log(2 << 0)

	var val int
	val |= 1 << 0
	val |= 1 << 1
	t.Log("contains:", val&(1<<0) != 0)
	t.Log("contains:", val&(1<<1) != 0)
}

// Demo: bytes, string

func TestReuseBytes(t *testing.T) {
	b := []byte("hello")
	t.Log(string(b))

	b = b[:0] // reuse bytes
	t.Log("size:", len(b))
	b = append(b, []byte("foo")...)
	t.Log(string(b))
}

func TestRegexpFindIndex(t *testing.T) {
	reg := regexp.MustCompile("mtime")
	pos := reg.FindIndex([]byte("response.order[0].mtime=1697202124"))
	if len(pos) == 2 {
		t.Logf("pos: [start=%d,end=%d]", pos[0], pos[1])
	}
}

func TestStrSplitByMultiSpace(t *testing.T) {
	str := "one_space two_space  three_space   end"
	fields := strings.Fields(str)
	t.Log("split fields:", fields)

	reg, err := regexp.Compile(`\s+`)
	if err != nil {
		t.Fatal(err)
	}
	fields = reg.Split(str, -1)
	t.Log("split fields by regexp:", fields)
}

func TestStrMultiReplace(t *testing.T) {
	replace := strings.NewReplacer(" ", "", `\n`, "", `\t`, "")

	str := `{\t"name": "foo",\n\t"age": 31,\n\t"skills:": ["java", "golang"]}`
	result := replace.Replace(str)
	t.Log("result:", result)
}

// Demo: ref

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

func TestRefUpdateForSlice(t *testing.T) {
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

func TestRefUpdateForMap(t *testing.T) {
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

// Demo: struct

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

// Demo: copy of struct

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

// Demo: goroutine

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
