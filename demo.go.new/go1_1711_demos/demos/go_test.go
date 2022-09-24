package demos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

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

func TestStringEqual(t *testing.T) {
	res := strings.EqualFold("foo", "Foo")
	t.Log("result:", res)
}

func TestRelativePath(t *testing.T) {
	dst := filepath.Join(os.Getenv("HOME"), "Workspaces/zj_repos/zj_new_go_project/demo.go.new/go1_1711_demos")
	relPath, err := filepath.Rel(os.Getenv("HOME"), dst)
	assert.NoError(t, err)
	t.Log("rel path:", relPath)
}

func TestIOReadCloser(t *testing.T) {
	// 从 request.body reader 中读出请求数据后，使用 io.NopCloser 还原 request.body reader
	r := strings.NewReader("io read closer test")
	rc := io.NopCloser(r)
	defer rc.Close()

	s, err := io.ReadAll(rc)
	assert.NoError(t, err)
	t.Log("read:", string(s))
}

func TestURLDecode(t *testing.T) {
	path := "/nice%20ports%2C/Tri%6Eity.txt%2ebak"
	res, err := url.PathUnescape(path)
	assert.NoError(t, err)
	t.Log("decode path:", res)
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

func TestStructToMap(t *testing.T) {
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

// Demo: error

func TestWrapError(t *testing.T) {
	err := fmt.Errorf("raw error")
	wErr := fmt.Errorf("wrapped error: %w", err)
	t.Log("error:", wErr)
	t.Log("error is:", errors.Is(wErr, err))
	t.Log("unwrap error:", errors.Unwrap(wErr))
}

func TestVerifyErrorType(t *testing.T) {
	err := &os.PathError{
		Path: "/tmp/test",
		Op:   "access",
		Err:  errors.New("path not exist"),
	}
	wErr := fmt.Errorf("wrapped error: %w", err)
	t.Log("error:", wErr)

	if _, ok := wErr.(*os.PathError); ok {
		t.Log("wrapped error is os.PathError")
	}

	var p *os.PathError
	if errors.As(wErr, &p) {
		t.Log("wrapped error as os.PathError")
	}
}

type CustomError struct {
	content string
}

func (e *CustomError) Error() string {
	return e.content
}

func TestCustomError(t *testing.T) {
	err := &CustomError{
		content: "system exception",
	}
	wErr := fmt.Errorf("wrapped error: %w", err)
	t.Log("error:", wErr)

	var tErr *CustomError
	t.Log("error as:", errors.As(wErr, &tErr))
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

func TestSliceParamRef(t *testing.T) {
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

// Demo: sort slice of structs

type person struct {
	name string
	age  int
	sex  string
}

type sortByPersonFields []person

func (e sortByPersonFields) Swap(i, j int) { e[i], e[j] = e[j], e[i] }
func (e sortByPersonFields) Len() int      { return len(e) }
func (e sortByPersonFields) Less(i, j int) bool {
	eleX := e[i]
	eleY := e[j]
	if eleX.sex != eleY.sex {
		return eleX.sex < eleY.sex
	}
	if eleX.age != eleY.age {
		return eleX.age < eleY.age
	}
	return eleX.name < eleY.name
}

func TestSortSliceOfStructs01(t *testing.T) {
	// sort by: sex asc, age asc, name asc
	persons := []person{
		{name: "zh", age: 24, sex: "female"},
		{name: "foo", age: 30, sex: "male"},
		{name: "yx", age: 36, sex: "female"},
		{name: "jx", age: 27, sex: "female"},
		{name: "zht", age: 24, sex: "female"},
		{name: "bar", age: 33, sex: "male"},
		{name: "ja", age: 24, sex: "female"},
	}

	sort.Sort(sortByPersonFields(persons))
	for _, p := range persons {
		log.Println("sorted persons:", p.sex, p.age, p.name)
	}
}

func TestSortSliceOfStructs02(t *testing.T) {
	// sort by: sex desc, age asc, name desc
	persons := []person{
		{name: "zh", age: 24, sex: "female"},
		{name: "foo", age: 30, sex: "male"},
		{name: "yx", age: 36, sex: "female"},
		{name: "jx", age: 27, sex: "female"},
		{name: "zht", age: 24, sex: "female"},
		{name: "bar", age: 33, sex: "male"},
		{name: "ja", age: 24, sex: "female"},
	}

	sort.Slice(persons, func(i, j int) bool {
		eleX := persons[i]
		eleY := persons[j]
		if eleX.sex != eleY.sex {
			return eleX.sex > eleY.sex
		}
		if eleX.age != eleY.age {
			return eleX.age < eleY.age
		}
		return eleX.name > eleY.name
	})

	for _, p := range persons {
		log.Println("sorted persons:", p.sex, p.age, p.name)
	}
}

// Demo: group by requests (type dict)

type dict map[string]interface{}

func groupByRequests(results map[string]int, reqs []dict) error {
	for {
		if len(reqs) == 0 {
			return nil
		}
		if len(reqs) == 1 {
			curReq := reqs[0]
			b, err := json.Marshal(curReq)
			if err != nil {
				return err
			}
			results[string(b)] = 1
			return nil
		}

		count := 1
		curReq := reqs[0]
		notMatchedReqs := make([]dict, 0, 16)
		for _, req := range reqs[1:] {
			if isDictEqualByFields(curReq, req) {
				count += 1
			} else {
				notMatchedReqs = append(notMatchedReqs, req)
			}
		}
		b, err := json.Marshal(curReq)
		if err != nil {
			return err
		}
		results[string(b)] = count

		reqs = notMatchedReqs
	}
}

// isDictEqualByFields only compares dict fields of 1st level.
func isDictEqualByFields(src, dst interface{}) bool {
	srcDict, ok := src.(dict)
	if !ok {
		return false
	}
	dstDict, ok := dst.(dict)
	if !ok {
		return false
	}

	if len(srcDict) != len(dstDict) {
		return false
	}

	for k := range srcDict {
		if _, ok := dstDict[k]; !ok {
			return false
		}
	}
	return true
}

func TestGroupByRequests(t *testing.T) {
	reqs := []dict{
		{"uid": 1221654048, "name": "ab"},
		{"uid": 1221654049, "name": "cd"},
		{"sp_uid": 1221724968},
		{"uid": 1221654048, "otp_token": "111111"},
		{"uid": 1221654050, "name": "xy"},
		{"uid": 1221654051, "name": "xz"},
		{"sp_uid": 1221724967},
	}

	results := make(map[string]int, 4)
	err := groupByRequests(results, reqs)
	assert.NoError(t, err)
	for req, count := range results {
		t.Logf("%s:%d", req, count)
	}
	t.Log("groupby done")
}

// Demo: goroutine

func TestSelectCase(t *testing.T) {
	ch := make(chan int, 1)
	go func() {
		time.Sleep(2 * time.Second)
		ch <- 1
		close(ch)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		t.Log(ctx.Err())
	case val, ok := <-ch:
		t.Log("chan value:", ok, val)
	}
	t.Log("done")
}

func TestRunBatchByGoroutine(t *testing.T) {
	resultCh := make(chan int)
	errCh := make(chan error)
	defer func() {
		close(resultCh)
		close(errCh)
	}()

	go func(resultCh chan int, errCh chan error) {
		for i := 0; i < 10; i++ {
			if i == 11 {
				errCh <- fmt.Errorf("invalid num")
			}
			resultCh <- i
			time.Sleep(time.Second)
		}
		// NOTE: size of resultCh should be 0, make sure all results are handle before return
		errCh <- nil
	}(resultCh, errCh)

outer:
	for {
		select {
		case result := <-resultCh:
			if result%2 == 1 {
				continue
			}
			t.Log("result:", result)
		case err := <-errCh:
			if err != nil {
				t.Log("err:", err)
			}
			break outer
		}
	}
	t.Log("done")
}

func TestGoroutineExit(t *testing.T) {
	// NOTE: sub goroutine is still running when root goroutine exit
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	retCh := make(chan struct{})
	go func() {
		// context here, make sure sub goroutine is cancelled when root goroutine exit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		go func() {
			tick := time.Tick(time.Second)
			for {
				select {
				case <-ctx.Done():
					fmt.Println("sub goroutine:", ctx.Err())
					return
				case <-tick:
					fmt.Println("sub goroutine run...")
				}
			}
		}()
		for i := 0; i < 5; i++ {
			time.Sleep(time.Second)
			fmt.Println("root goroutine run...")
		}

		// <-retCh
		close(retCh)
	}()

	t.Log("main wait...")
	<-retCh
	t.Log("root goroutine finish")
	time.Sleep(3 * time.Second)
	t.Log("main done")
}
