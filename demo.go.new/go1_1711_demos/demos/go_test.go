package demos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// run multiple tests:
// go test -timeout 30s -run "TestChar|TestStructToMap" go1_1711_demo/demos -v -count=1
func TestMain(m *testing.M) {
	fmt.Println("test before")
	code := m.Run()
	fmt.Printf("return code: %d\n", code)
	fmt.Println("test after")
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

func TestURLDecode(t *testing.T) {
	path := "/nice%20ports%2C/Tri%6Eity.txt%2ebak"
	res, err := url.PathUnescape(path)
	assert.NoError(t, err)
	t.Log("decode path:", res)
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
		{"uid": 1221654048},
		{"uid": 1221654049},
		{"sp_uid": 1221724968},
		{"uid": 1221654048, "otp_token": "111111"},
		{"uid": 1221654050},
		{"uid": 1221654051},
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
