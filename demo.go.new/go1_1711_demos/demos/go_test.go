package demos

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

//
// run multiple tests:
// go test -timeout 30s -run "TestChar|TestStructToMap" go1_1711_demo/demos -v -count=1
//

func TestMain(m *testing.M) {
	fmt.Println("test before")
	code := m.Run()
	fmt.Printf("return code: %d\n", code)
	fmt.Println("test after")
}

// Demo: 内存对齐

type s1 struct {
	a int8
	b int16
	c int32
}

type s2 struct {
	a int8
	c int32
	b int16
}

func TestStructSize(t *testing.T) {
	// 在对内存特别敏感的结构体的设计上，我们可以通过调整字段的顺序，将字段宽度从小到大由上到下排列，来减少内存的占用
	t.Log("s1 size:", unsafe.Sizeof(s1{})) // 8
	t.Log("s2 size:", unsafe.Sizeof(s2{})) // 12
}

// Demo: string

func TestStringEqual(t *testing.T) {
	res := strings.EqualFold("foo", "Foo")
	t.Log("result:", res)
}

func TestStringBuilder(t *testing.T) {
	s1, s2, s3 := "foo|", "bar|", "baz"
	var builder strings.Builder
	builder.Grow(9)
	_, err := builder.WriteString(s1)
	assert.NoError(t, err)
	_, err = builder.WriteString(s2)
	assert.NoError(t, err)
	_, err = builder.WriteString(s3)
	assert.NoError(t, err)
	t.Log("results:", builder.String())
}

func TestChar(t *testing.T) {
	c := fmt.Sprintf("%c", 119)
	t.Logf("str=%s, len=%d", c, len(c))

	c = fmt.Sprintf("%c", 258)
	t.Logf("str=%s, len=%d", c, len(c))

	r := rune('中')
	t.Logf("char=%c, d=%d", r, r)
	c = fmt.Sprintf("%c", r)
	t.Logf("str=%s, len=%d", c, len(c))

	c = fmt.Sprintf("%c", 20132)
	t.Logf("str=%s, len=%d", c, len(c))

	s := "中cn"
	t.Logf("size=%d", len(s))
}

func TestURLDecode(t *testing.T) {
	path := "/nice%20ports%2C/Tri%6Eity.txt%2ebak"
	res, err := url.PathUnescape(path)
	assert.NoError(t, err)
	t.Log("decode path:", res)
}

// Demo: slice 初始值判断

type TestSliceHolder struct {
	data []string
}

func TestSliceCheck(t *testing.T) {
	holder := TestSliceHolder{}
	if holder.data == nil {
		t.Log("slice is <nil>")
	}
	if len(holder.data) == 0 {
		t.Log("slice len:", len(holder.data))
	}

	holder.data = make([]string, 3)
	t.Log("slice len:", len(holder.data))
}

// Demo: iterator

func TestArrayIterator(t *testing.T) {
	s := [3]int{1, 2, 3}
	for i := 0; i < len(s); i++ {
		s[i]++
		t.Logf("elem: %p, %v\n", &s[i], s[i])
	}
	t.Log(s)

	// copy value of s to v
	for _, v := range s {
		v++
		t.Logf("elem: %p, %v\n", &v, v)
	}
	t.Log(s)
}

func TestSliceIterator(t *testing.T) {
	s := []int{1, 2, 3}
	for i := 0; i < len(s); i++ {
		s[i]++
		t.Logf("elem: %p, %v\n", &s[i], s[i])
	}
	t.Log(s)

	// copy value of s to v
	for _, v := range s {
		v++
		t.Logf("elem: %p, %v\n", &v, v)
	}
	t.Log(s)
}

func TestMapIterator(t *testing.T) {
	m := map[string]int{
		"one":   0,
		"two":   1,
		"three": 2,
	}
	for k := range m {
		m[k]++
	}
	t.Logf("%+v", m)

	for k, v := range m {
		v++
		t.Logf("key:%p,%5s | value:%p,%d", &k, k, &v, v)
	}
	t.Logf("%+v", m)
}

func TestTimeTickIterator(t *testing.T) {
	tick := time.NewTicker(time.Second)
	go func() {
		time.Sleep(5 * time.Second)
		tick.Stop()
	}()

	// here pending after tick stopped
	for x := range tick.C {
		t.Log("run by second:", x.Second())
	}
	t.Log("tick stopped")
}

// Demo: param ref

func TestArrayParamRef(t *testing.T) {
	// array elem is pass by cpoied value
	updateSlice := func(s [3]int) {
		for i := 0; i < 3; i++ {
			s[i]++
		}
		t.Logf("#2: %p, %p, %v", &s, &s[0], s)
	}

	s := [3]int{1, 2, 3}
	t.Logf("#1: %p, %p, %v", &s, &s[0], s)
	updateSlice(s)
	t.Logf("#3: %p, %p, %v", &s, &s[0], s)
}

func TestSliceParamRef01(t *testing.T) {
	// slice elem is pass by ref
	updateSlice := func(s []int) {
		for i := 0; i < len(s); i++ {
			s[i]++
		}
		t.Logf("#2: %p, %p, %v", &s, &s[0], s)
	}

	s := []int{1, 2, 3}
	t.Logf("#1: %p, %p, %v", &s, &s[0], s)
	updateSlice(s)
	t.Logf("#2: %p, %p, %v", &s, &s[0], s)
}

func TestSliceParamRef02(t *testing.T) {
	prettyPrint := func(s []int, prefix string) {
		content := fmt.Sprintf("%s: s=%v,len=%d,cap=%d", prefix, s, len(s), cap(s))
		t.Log(content)
	}

	// 若在函数中对该切片进行追加（append）且追加后的切片大小不超过其原始容量，此时修改切片中已有的元素，则修改会同步到实参切片中，而追加不会同步到实参切片中。
	updateSliceWithinCap := func(s []int) {
		s = append(s, 10)
		s[0]++
		prettyPrint(s, "update within cap, inner")
	}

	// 若在函数中对该切片进行追加且追加后的切片大小超过其原始容量，则修改不会同步到实参切片中，同时追加也不会同步到实参切片中。
	updateSliceOverCap := func(s []int) {
		s = append(s, 20)
		s = append(s, 21)
		s[0]++
		prettyPrint(s, "update over cap, inner")
	}

	s := make([]int, 1, 2)
	s[0] = 1
	prettyPrint(s, "s src:")

	updateSliceWithinCap(s)
	prettyPrint(s, "update within cap, outer")

	updateSliceOverCap(s)
	prettyPrint(s, "update over cap, outer")
}

func TestMapParamRef(t *testing.T) {
	// map value is pass by ref
	updateMap := func(m map[string]int) {
		for k := range m {
			m[k]++
		}
		t.Logf("#2: %p, %+v", &m, m)
	}

	m := map[string]int{
		"one":   0,
		"two":   1,
		"three": 2,
	}
	t.Logf("#1: %p, %+v", &m, m)
	updateMap(m)
	t.Logf("#3: %p, %+v", &m, m)
}

// Demo: struct inherit

type testParent struct {
	Name string
}

func (p testParent) Pprint() {
	fmt.Printf("name=%s\n", p.Name)
}

type testChild struct {
	testParent
	Age int
}

func (c testChild) Pstring() string {
	c.Pprint()
	return fmt.Sprintf("name=%s, age=%d\n", c.Name, c.Age)
}

func TestStructInherit(t *testing.T) {
	c := testChild{
		testParent: testParent{Name: "foo"},
		Age:        31,
	}
	c.Pprint()
	t.Log(c.Pstring())
}

// Demo: context

func TestContextWithValue(t *testing.T) {
	type person struct {
		name string
		age  int
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := &person{
		name: "foo",
		age:  31,
	}
	newCtx := context.WithValue(ctx, "key", p)
	time.Sleep(300 * time.Millisecond)

	p, ok := newCtx.Value("key").(*person)
	if !ok {
		t.Fatal("type error")
	}
	t.Logf("name=%s, age=%d", p.name, p.age)
}

func TestContinueInSelect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tick := time.Tick(time.Second)
	for i := 0; i < 10; i++ {
		t.Logf("run at: %d", i)
		select {
		case <-ctx.Done():
			t.Log("cancelled")
		case <-tick:
			if i%2 == 1 {
				continue
			}
			t.Logf("select at: %d", i)
		}
	}
	t.Log("done")
}
