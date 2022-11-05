package demos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

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

func TestBarrierByWg(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		i := i
		go func() {
			time.Sleep(time.Duration(i*10) * time.Millisecond)
			t.Logf("goroutine [%d] before wait", i)
			wg.Done()
			wg.Wait()
			t.Logf("goroutine [%d] after wait", i)
		}()
	}

	// print "after wait" after all "before wait"
	time.Sleep(3 * time.Second)
	t.Log("barrier test done")
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

//
// Demo: 无锁数据结构 atomic
//

type Node struct {
	Value interface{}
	Next  *Node
}

// WithLockList 有锁单向链表
type WithLockList struct {
	Head *Node
	mux  sync.Mutex
}

// Push 将元素插入到链表的首部
func (l *WithLockList) Push(v interface{}) {
	l.mux.Lock()
	defer l.mux.Unlock()
	n := &Node{
		Value: v,
		Next:  l.Head,
	}
	l.Head = n
}

func (l *WithLockList) String() string {
	s := ""
	cur := l.Head
	for cur != nil {
		if s != "" {
			s += ","
		}
		s += fmt.Sprintf("%v", cur.Value)
		cur = cur.Next
	}
	return s
}

func TestWriteWithLockList(t *testing.T) {
	var g errgroup.Group
	l := &WithLockList{}
	for i := 0; i < 10; i++ {
		i := i
		g.Go(func() error {
			l.Push(i)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
	t.Log(l)
}

// LockFreeList 无锁单向链表
type LockFreeList struct {
	Head atomic.Value
}

func (l *LockFreeList) Push(v interface{}) {
	for {
		head := l.Head.Load()
		headNode, _ := head.(*Node)
		n := &Node{
			Value: v,
			Next:  headNode,
		}
		if l.Head.CompareAndSwap(head, n) {
			return
		}
	}
}

func (l *LockFreeList) String() string {
	s := ""
	cur := l.Head.Load().(*Node)
	for cur != nil {
		if s != "" {
			s += ","
		}
		s += fmt.Sprintf("%v", cur.Value)
		cur = cur.Next
	}
	return s
}

func TestWriteLockFreeList(t *testing.T) {
	var g errgroup.Group
	l := &LockFreeList{}
	for i := 0; i < 10; i++ {
		i := i
		g.Go(func() error {
			l.Push(i)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
	t.Log(l)
}

//
// Demo: sync.Cond
//

func TestSyncCond(t *testing.T) {
	done := false
	write := func(name string, c *sync.Cond) {
		t.Log(name, "starts writing")
		time.Sleep(time.Second)
		done = true
		t.Log(name, "wakes all")
		c.Broadcast()
	}

	read := func(name string, c *sync.Cond) {
		c.L.Lock()
		defer c.L.Unlock()
		for !done {
			c.Wait()
		}
		t.Log(name, "starts reading")
	}

	cond := sync.NewCond(&sync.Mutex{})
	go read("read1", cond)
	go read("read2", cond)
	go read("read3", cond)

	write("writer", cond)
	time.Sleep(3 * time.Second)
	t.Log("sync.cond test done")
}

//
// Demo: sync.Pool 使用
//

type SyncPoolTestStudent struct {
	Name   string
	Age    int32
	Remark [1024]byte
}

func TestSyncPool(t *testing.T) {
	studentPool := sync.Pool{
		New: func() interface{} {
			return new(SyncPoolTestStudent)
		},
	}

	buf := []byte(`{"name":"foo","age":31}`)
	s := studentPool.Get().(*SyncPoolTestStudent)
	if err := json.Unmarshal(buf, s); err != nil {
		t.Fatal(err)
	}
	studentPool.Put(s)

	s = studentPool.Get().(*SyncPoolTestStudent)
	t.Logf("student: name=%s, age=%d", s.Name, s.Age)
}

func BenchmarkSyncPool(b *testing.B) {
	bufPool := sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}

	data := make([]byte, 10000)
	for n := 0; n < b.N; n++ {
		buf := bufPool.Get().(*bytes.Buffer)
		buf.Write(data)
		buf.Reset()
		bufPool.Put(buf)
	}
}

//
// Demo: data race condition
//

func TestSliceSafeAppend(t *testing.T) {
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
