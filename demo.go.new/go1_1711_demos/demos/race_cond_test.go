package demos

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

//
// Demo: data race condition
//

func TestArrayCopy(t *testing.T) {
	// 元素为值传递
	updateSlice := func(s [3]int) {
		for i := 0; i < 3; i++ {
			s[i]++
		}
		t.Logf("dst: %p, %p, %v", &s, &s[0], s)
	}

	s := [3]int{1, 2, 3}
	updateSlice(s)
	t.Logf("src: %p, %p, %v", &s, &s[0], s)
}

func TestSliceCopy(t *testing.T) {
	// 元素为引用传递
	updateSlice := func(s []int) {
		for i := 0; i < len(s); i++ {
			s[i]++
		}
		t.Logf("dst: %p, %p, %v", &s, &s[0], s)
	}

	s := []int{1, 2, 3}
	t.Logf("src: %p, %p, %v", &s, &s[0], s)
	updateSlice(s)
}

func TestMapCopy(t *testing.T) {
	// kv 为引用传递
	updateMap := func(m map[string]int) {
		for k, v := range m {
			v++
			t.Logf("dst: %p, %p, %s:%d", &k, &v, k, v)
		}
	}

	m := map[string]int{
		"one":   0,
		"two":   1,
		"three": 2,
	}
	for k, v := range m {
		v++
		t.Logf("\nsrc: %p, %p, %s:%d", &k, &v, k, v)
	}
	updateMap(m)
}

func TestSliceAppend(t *testing.T) {
	var mutex sync.Mutex
	safeAppend := func(name string, names []string) {
		mutex.Lock()
		names[0] += "_x"
		names = append(names, name) // "names" ref to new slice
		fmt.Println("dst:", names)
		mutex.Unlock()
	}

	names := []string{"foo"}
	for _, name := range strings.Split("ab|cd|ef", "|") {
		go safeAppend(name, names)
	}
	time.Sleep(time.Second)
	t.Logf("src: %v", names)
}

func TestRaceCondition01(t *testing.T) {
	type person struct {
		name string
		age  int
	}

	persons := []person{
		{name: "foo", age: 31},
		{name: "bar", age: 40},
	}

	for _, p := range persons {
		local := p // use local var
		go func() {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("name=%s,age=%d\n", local.name, local.age)
		}()
	}
	time.Sleep(time.Second)
	t.Log("done")
}

func NamedReturnCallee(flag bool) (result int) {
	result = 10
	if flag {
		return
	}

	local := result
	go func() {
		for i := 0; i < 3; i++ {
			fmt.Println(local, result)
		}
	}()
	return 20
}

func TestNamedReturnCallee(t *testing.T) {
	ret := NamedReturnCallee(false)
	t.Log("result:", ret)
	time.Sleep(time.Second)
	t.Log("done")
}
