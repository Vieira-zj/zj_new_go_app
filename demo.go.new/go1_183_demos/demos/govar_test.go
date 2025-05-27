package demos_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/bits"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// Demo: Bit

func TestBitsOps(t *testing.T) {
	t.Run("bit move op", func(t *testing.T) {
		t.Log(0 << 2)  // 0
		t.Log(16 << 1) // num*2
		t.Log(14 >> 1) // num/2
	})

	t.Run("bit or op", func(t *testing.T) {
		t.Log(1 << 1)
		t.Log(1 << 2)

		var val int
		val |= 1 << 1
		val |= 1 << 2
		t.Log("contains:", val&(1<<1) != 0)
		t.Log("contains:", val&(1<<2) != 0)
	})

	t.Run("bits util", func(t *testing.T) {
		// 1 byte: [0,255]
		num := uint16(259)
		t.Log("bit count:", bits.OnesCount(uint(num)))

		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, num)
		t.Log("bytes:", b) // [0000 0001], [0000 0011]
	})
}

// Demo: Number

func TestHexToDecimal(t *testing.T) {
	val := 0xff
	result := strconv.FormatInt(int64(val), 10)
	t.Logf("decimal result: %d, %s", int64(val), result)
}

// Demo: Bytes, String

func TestBytesCompare(t *testing.T) {
	t.Run("bytes compare", func(t *testing.T) {
		result := bytes.Compare([]byte("abc"), []byte("abx"))
		t.Log("result:", result)
	})

	t.Run("string compare", func(t *testing.T) {
		result := strings.Compare("abx", "abc")
		t.Log("result:", result)
	})
}

func TestReuseBytes(t *testing.T) {
	b := []byte("hello")
	t.Log(string(b))

	b = b[:0] // reuse bytes
	t.Logf("len=%d, cap=%d", len(b), cap(b))

	b = append(b, []byte("foo")...)
	t.Log(string(b))
	t.Logf("len=%d, cap=%d", len(b), cap(b))
}

func TestStrOps(t *testing.T) {
	t.Run("string quota", func(t *testing.T) {
		str := "hello"
		t.Logf("quota str: %s, %q", strconv.Quote(str), str)
	})

	t.Run("string cut", func(t *testing.T) {
		str := "foo|hello world"
		before, after, ok := strings.Cut(str, "|")
		if ok {
			t.Logf("before=%s, after=%s", before, after)
		}
	})
}

func TestStrSplitByMultiSpace(t *testing.T) {
	str := "one_space two_space  three_space   end"
	t.Run("strings fields", func(t *testing.T) {
		fields := strings.Fields(str)
		t.Log("split fields:", fields)
	})

	t.Run("regexp split", func(t *testing.T) {
		reg, err := regexp.Compile(`\s+`)
		if err != nil {
			t.Fatal(err)
		}
		fields := reg.Split(str, -1)
		t.Log("split fields by regexp:", fields)
	})
}

func TestStrMultiReplace(t *testing.T) {
	replace := strings.NewReplacer(" ", "", `\n`, "", `\t`, "")

	str := `{\t"name": "foo",\n\t"age": 31,\n\t"skills:": ["java", "golang"]}`
	result := replace.Replace(str)
	t.Log("result:", result)
}

// Demo: Slice

func TestArrayOps(t *testing.T) {
	t.Run("array loop", func(t *testing.T) {
		arr := [3]int{1, 2, 3}
		for i := 0; i < len(arr); i++ {
			t.Logf("arr[%d]: %d", i, arr[i])
		}

		for i, v := range arr {
			t.Logf("arr[%d]: %d", i, v)
		}
	})
}

func TestSliceInit(t *testing.T) {
	t.Run("init slice by append", func(t *testing.T) {
		var s []string
		s = append(s, strings.Split("hello", "")...)
		t.Log("slice:", s)
	})

	t.Run("init slice as map value", func(t *testing.T) {
		m := make(map[byte][]string)
		for _, s := range []string{
			"a1", "a2", "apple",
			"b1", "b2", "banana",
		} {
			b := s[0]
			m[b] = append(m[b], s)
		}

		for k, v := range m {
			t.Logf("key=%s, value=%v", string(k), v)
		}
	})

	t.Run("init slice with index", func(t *testing.T) {
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
	})
}

func TestSliceToArray(t *testing.T) {
	t.Run("slice to array for Go 1.17", func(t *testing.T) {
		var a [3]int
		s := []int{0, 1, 2, 3, 4, 5}
		a = *(*[3]int)(s[:3])
		t.Log("array:", a)
	})

	t.Run("slice to array for Go 1.20", func(t *testing.T) {
		var a [3]int
		s := []int{0, 1, 2, 3, 4, 5}
		a = [3]int(s[:3])
		t.Log("array:", a)
	})
}

func TestSliceSubOps(t *testing.T) {
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

	t.Run("sub slice case1", func(t *testing.T) {
		s1 := s[:0]
		t.Logf("len=%d, cap=%d: %v", len(s1), cap(s1), s1)
	})

	t.Run("sub slice case2", func(t *testing.T) {
		s1 := s[:2]
		t.Logf("len=%d, cap=%d: %v", len(s1), cap(s1), s1)
	})

	t.Run("sub slice case3", func(t *testing.T) {
		s1 := s[2:]
		t.Logf("len=%d, cap=%d: %v", len(s1), cap(s1), s1)
	})

	t.Run("sub slice case4", func(t *testing.T) {
		// len = 3-1 = 2
		// cap = s.len-1 = 3
		s1 := s[1:3]
		t.Logf("len=%d, cap=%d: %v", len(s1), cap(s1), s1)
	})

	t.Run("sub slice case5 (set cap manual)", func(t *testing.T) {
		// len = 3-1 = 2
		// cap = 4-1 = 3
		s1 := s[1:3:4]
		t.Logf("len=%d, cap=%d: %v", len(s1), cap(s1), s1)
	})
}

func TestSliceAppend(t *testing.T) {
	t.Run("slice append without scale", func(t *testing.T) {
		s1 := make([]int, 1, 2)
		s1[0] = -1

		s2 := append(s1, 2)
		s2[0] = -2

		s3 := append(s1, 3)
		s3[0] = -3

		// s,s2,s3 share the same address of array
		for i, s := range [][]int{s1, s2, s3} {
			t.Logf("s%d: s=%v, ptr=%p, elem=%p", i+1, s, s, &s[0])
		}
	})

	t.Run("slice append with scale", func(t *testing.T) {
		s1 := make([]int, 1, 2)
		s1[0] = -1

		s2 := append(s1, 2)
		s2[0] = -2

		// scale
		s3 := append(s2, 3)
		s3[0] = -3

		// after scale, and get new address of slice
		for i, s := range [][]int{s1, s2, s3} {
			t.Logf("s%d: s=%v, ptr=%p, elem=%p", i+1, s, s, &s[0])
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

func TestSliceGrow(t *testing.T) {
	s := make([]byte, 2, 10)
	s[0] = 0
	s[1] = 1
	t.Logf("init slice: %v, ptr=%p", s, &s[0])

	l := len(s)

	t.Run("slice copy without grow", func(t *testing.T) {
		sub := []byte{2, 3}
		copy(s[l:], sub)
		t.Logf("copied slice: %v, ptr=%p", s, &s[0]) // slice unchanged
	})

	t.Run("slice grow and copy", func(t *testing.T) {
		sub := []byte{2, 3, 4}
		s = s[:l+len(sub)]
		t.Logf("grow slice: %v, ptr=%p", s, &s[0])

		copy(s[l:], sub)
		t.Logf("copied slice: %v, ptr=%p", s, &s[0])
	})

	t.Run("slice append", func(t *testing.T) {
		s = append(s, 5)
		t.Logf("append slice: %v, ptr=%p", s, &s[0])
	})

	t.Run("slice append with scale", func(t *testing.T) {
		for i := 6; i < 13; i++ {
			s = append(s, byte(i))
		}
		t.Logf("append slice: %v, ptr=%p, cap=%d", s, &s[0], cap(s))
	})

	t.Run("slice grow with scale", func(t *testing.T) {
		s := make([]byte, 2, 3)
		s[0] = 0
		s[1] = 1
		t.Logf("slice: %v, ptr=%p, elem=%p", s, s, &s[0])

		// panic: slice bounds out of range
		// s = s[:len(s)+2]
		// t.Log("slice:", s)
	})
}

// Demo: Map

func TestMapCap(t *testing.T) {
	m := make(map[int]string, 2)
	m[1] = "one"
	t.Logf("len=%d", len(m))

	for k, v := range m {
		t.Logf("key=%d, value=%s", k, v)
	}
}

func TestMapKeyByStruct(t *testing.T) {
	type num struct {
		id    int
		value string
	}

	t.Run("struct as map key", func(t *testing.T) {
		three := num{id: 3, value: "three"}
		four := num{id: 4, value: "four"}

		m := map[num]string{
			three: "3:three",
			four:  "4:four",
		}

		t.Log("iterator:")
		for k, v := range m {
			t.Log(k.id, k.value, v)
		}

		t.Log("check key [one] exist:")
		if v, ok := m[three]; ok {
			t.Log("found", v)
		} else {
			t.Fatal("not found")
		}

		newThree := num{id: 3, value: "three"}
		t.Logf("get with new existing key: id=%d, value=%s", newThree.id, newThree.value)
		if v, ok := m[newThree]; ok {
			t.Log("found", v)
		} else {
			t.Log("not found")
		}
	})

	t.Run("struct ptr as map key", func(t *testing.T) {
		one := &num{id: 1, value: "one"}
		two := &num{id: 2, value: "two"}

		// use address as map key
		m := map[*num]string{
			one: "1:one",
			two: "2:two",
		}

		t.Log("iterator:")
		for k, v := range m {
			t.Log(k.id, k.value, v)
		}

		t.Log("check key [one] exist:")
		if v, ok := m[one]; ok {
			t.Log("found", v)
		} else {
			t.Fatal("not found")
		}

		newOne := &num{id: 1, value: "one"}
		t.Logf("get with new existing key: id=%d, value=%s", newOne.id, newOne.value)
		if v, ok := m[newOne]; ok {
			t.Log("found", v)
		} else {
			t.Log("not found")
		}
	})
}

// Demo: Struct

type testStruct struct {
	id   int
	name string
}

func (s testStruct) String() string {
	return fmt.Sprintf("id=%d, name=%s", s.id, s.name)
}

func TestUpdateStruct(t *testing.T) {
	s := testStruct{
		id:   1,
		name: "foo",
	}
	s1 := s

	s.id = 2
	s.name = "bar"
	s2 := s
	t.Log("src:", s1)
	t.Log("dest:", s2)
}

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

func TestUpdateStudentStruct(t *testing.T) {
	s1 := NewTestStudent("foo", 13)
	s1.addTag("p1")
	s1.addScore("en", 91)
	t.Logf("s1 [%p]: %s", &s1, s1.String())

	s2 := updateStudent(s1)
	t.Log("after update")
	t.Log("s1:", s1.String())
	t.Log("s2:", s2.String())
}

func TestUpdateStudentRef(t *testing.T) {
	s := NewTestStudent("foo", 13)
	s.addTag("p1")
	s.addScore("en", 91)
	t.Logf("s [%p]: %s", &s, s.String())

	updateStudentRef(&s)
	t.Log("after update")
	t.Logf("s [%p]: %s", &s, s.String())
}

// Demo: Copy of Struct

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

// Demo: Abstract Class

type IPerson interface {
	GetName() string
	Say(string)
}

type PersonBase struct{}

func (b PersonBase) GetName() string {
	return "default"
}

func (b PersonBase) Say(msg string) {
	fmt.Println("hello:", msg)
}

type AbstractPerson struct {
	PersonBase
}

// GetName overwrites the same method in PersonBase.
func (b AbstractPerson) Say(msg string) {
	fmt.Println("receive message:", msg)
}

func TestAbstractStruct(t *testing.T) {
	run := func(p IPerson) {
		t.Log("name:", p.GetName())
		p.Say("this is abstract class test")
	}

	p := AbstractPerson{}
	run(p)
}

// Demo: Atomic

func TestAtomicSumUpBySelfLoop(t *testing.T) {
	var (
		val atomic.Int32
		wg  sync.WaitGroup
	)

	limit := 50

	val.Store(1)
	wg.Add(limit)
	for i := 0; i < limit; i++ {
		go func(idx int) {
			defer wg.Done()
			// self-loop
			for {
				tmp := val.Load()
				fmt.Printf("goroutine [%d] load value: %d\n", idx, tmp)
				if ok := val.CompareAndSwap(tmp, tmp+1); ok {
					return
				} else {
					fmt.Printf("goroutine [%d] compare '%d' failed and retry\n", idx, tmp)
				}
			}
		}(i)
	}

	wg.Wait()
	t.Log("result:", val.Load())
}
