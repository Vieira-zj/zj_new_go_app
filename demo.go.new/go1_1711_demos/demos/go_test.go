package demos

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
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
	if err != nil {
		t.Fatal(err)
	}
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

// Demo: get func name and run by reflect

type caller func(string)

func sayHello(name string) {
	fmt.Println("Hello:", name)
}

func exec(c interface{}, params ...interface{}) {
	typeOf := reflect.TypeOf(c)
	fmt.Println("type:", typeOf.Kind())
	if typeOf.Kind() != reflect.Func {
		fmt.Println("not caller")
		return
	}

	// get func name
	valueOf := reflect.ValueOf(c)
	name := runtime.FuncForPC(valueOf.Pointer()).Name()
	pkgName, funcName := getFuncName(name)
	fmt.Printf("exec: pkg=%s, func=%s()\n", pkgName, funcName)

	// run func()
	paramValues := make([]reflect.Value, 0, len(params))
	for _, param := range params {
		paramValues = append(paramValues, reflect.ValueOf(param))
	}
	valueOf.Call(paramValues)

	_, ok := valueOf.Interface().(caller)
	fmt.Println("is caller:", ok)
}

func getFuncName(fullName string) (pkgName, funcName string) {
	items := strings.Split(fullName, ".")
	return items[0], items[1]
}

func TestGetFuncNameByReflect(t *testing.T) {
	exec(sayHello, "foo")
	t.Log("done")
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
