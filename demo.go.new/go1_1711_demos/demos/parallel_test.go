package demos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
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
		if ok {
			t.Log("chan value:", ok, val)
		}
	}
	t.Log("done")
}

func TestNotifyChan(t *testing.T) {
	sig := make(chan *string)

	tick := time.NewTicker(time.Second)
	defer tick.Stop()

	go func() {
		time.Sleep(3 * time.Second)
		t.Log("close chan")
		close(sig)
	}()

	go func() {
		time.Sleep(2 * time.Second)
		t.Log("send exit sig")
		sig <- nil
	}()

outer:
	for {
		select {
		case <-tick.C:
			t.Log("wait...")
		case _, ok := <-sig:
			if !ok {
				t.Log("chan closed")
			} else {
				t.Log("exit by sig")
			}
			break outer
		}
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

func TestAtomicWrappedMap(t *testing.T) {
	// panic: comparing uncomparable type map[string]int
	// uncomparable: map, slice, func
	m := map[string]int{
		"count1": 0,
		"count2": 0,
	}

	var val atomic.Value
	val.Store(m)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		i := i
		wg.Add(1)
		go func() {
			lval := val.Load()
			lmap, _ := lval.(map[string]int)
			key := fmt.Sprintf("count%d", i%2)
			for x := 0; x < 1000; x++ {
				lmap[key] += 1
			}
			val.CompareAndSwap(lval, lmap)
			wg.Done()
		}()
	}

	wg.Wait()
	for k, v := range m {
		t.Logf("%s=%d", k, v)
	}
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
	// NOTE: sub goroutine is still running when root goroutine exited
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

//
// Demo: singleflight
//

func getCacheBySingleflightDo(g *singleflight.Group, key string, id int) (string, error) {
	ret, err, shared := g.Do(key, func() (ret interface{}, err error) {
		fmt.Printf("[id=%d] run get cache\n", id)
		sleep := id
		if sleep > 3 {
			sleep = 3
		}
		time.Sleep(time.Duration(sleep) * time.Second)
		return id, nil
	})
	fmt.Printf("[id=%d] run get cache finish: shared=%v\n", id, shared)
	return fmt.Sprintf("%v", ret), err
}

func TestSingleflight(t *testing.T) {
	const (
		key   = "singleflight"
		count = 5
	)
	var (
		wg sync.WaitGroup
		g  singleflight.Group
	)

	wg.Add(count)
	// id 为 x 的请求率先发起了获取缓存，其他 4 个 goroutine 并不会去执行获取缓存的逻辑（相同的 key），而是等到 id 为 x 的请求取得结果后直接使用该结果
	for i := 0; i < count; i++ {
		go func(idx int) {
			defer wg.Done()
			fmt.Printf("[id=%d] ask cache\n", idx)
			val, err := getCacheBySingleflightDo(&g, key, idx)
			if err != nil {
				fmt.Println("error:", err)
			}
			log.Printf("[id=%d] get cache: key=%s,value=%s", idx, key, val)
		}(i)
	}
	wg.Wait()

	idx := 10
	fmt.Printf("[id=%d] ask cache\n", idx)
	val, err := getCacheBySingleflightDo(&g, key, idx)
	if err != nil {
		fmt.Println("error:", err)
	}
	log.Printf("[id=%d] get cache: key=%s,value=%s", idx, key, val)
	t.Log("singleflight test done")
}

func getCacheBySingleflightDoChan(g *singleflight.Group, key string, id int) (string, error) {
	retCh := g.DoChan(key, func() (ret interface{}, err error) {
		fmt.Printf("[id=%d] run get cache\n", id)
		sleep := id + 3
		if id == 10 {
			sleep = 1
		}
		fmt.Printf("[id=%d] sleep: %d sec\n", id, sleep)
		time.Sleep(time.Duration(sleep) * time.Second)
		return id, nil
	})

	// 注意：当没有设置超时时间，且第一个进入的 goroutine 执行被阻塞时，后面的 goroutine 都会被阻塞（相同的 key）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		g.Forget(key)
		return "-1", ctx.Err()
	case ret := <-retCh:
		fmt.Printf("[id=%d] run get cache finish: shared=%v\n", id, ret.Shared)
		return fmt.Sprintf("%v", ret.Val), ret.Err
	}
}

func TestSingleflightWithTimeout(t *testing.T) {
	const key = "singleflight"
	var (
		wg sync.WaitGroup
		g  singleflight.Group
	)

	// 如果第一个进入的 goroutine 执行时间大于 3s, 则 5 个 goroutine 都会报错 context deadline exceeded
	// 当小于 3s 时，5 个 goroutine 能正常返回
	for i := 0; i < 5; i++ {
		wg.Add(1)
		fmt.Printf("[id=%d] ask cache\n", i)
		go func(idx int) {
			defer wg.Done()
			val, err := getCacheBySingleflightDoChan(&g, key, idx)
			if err != nil {
				fmt.Printf("[id=%d] get cache error: %v\n", idx, err)
				return
			}
			log.Printf("[id=%d] get cache: key=%s, value=%s", idx, key, val)
		}(i)
	}
	wg.Wait()

	// 如果上 1 次调用还没有执行完成，且超时退出后也没有 Forget(key), 则 goroutine idx=10 会继续阻塞，然后报错 context deadline exceeded
	idx := 10
	val, err := getCacheBySingleflightDoChan(&g, key, idx)
	if err != nil {
		fmt.Printf("[id=%d] get cache error: %v\n", idx, err)
		return
	}
	log.Printf("[id=%d] get cache: key=%s, value=%s", idx, key, val)
	t.Log("singleflight test done")
}
