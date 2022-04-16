package demos

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

func demo01() string {
	return "demo01"
}

// demo02, context with cancel
func demo02() int {
	var (
		a       = 2
		b       = 4
		timeout = 3
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(time.Duration(timeout) * time.Second)
		cancel()
	}()

	ret := myAdd(ctx, a, b)
	fmt.Printf("\nCompute: %d+%d, result: %d\n", a, b, ret)
	return ret
}

/*
context with timeout
*/

func demo03() int {
	var (
		a       = 2
		b       = 4
		timeout = 5
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	ret := myAdd(ctx, a, b)
	fmt.Printf("\nCompute: %d+%d, result: %d\n", a, b, ret)
	return ret
}

func myAdd(ctx context.Context, a, b int) int {
	ret := 0
	for i := 0; i < a; i++ {
		fmt.Println("increase a")
		ret = myIncr(ret)
		select {
		case <-ctx.Done():
			fmt.Println("a: cancel incr()")
			return ret
		default:
		}
	}

	for i := 0; i < b; i++ {
		fmt.Println("increase b")
		ret = myIncr(ret)
		select {
		case <-ctx.Done():
			fmt.Println("b: cancel incr()")
			return ret
		default:
		}
	}
	return ret
}

func myIncr(x int) int {
	time.Sleep(time.Second)
	return x + 1
}

/*
context with value
*/

type ctxKey int

type ctxData struct {
	value string
}

// enum as context key
const (
	key1 ctxKey = iota
	key2
)

func demo04() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Duration(3)*time.Second))
	defer cancel()

	data := ctxData{
		value: "test",
	}
	ctx = context.WithValue(ctx, key1, "value1")
	ctx = context.WithValue(ctx, key2, &data)

	done := make(chan struct{})
	go func() {
		fmt.Println("key1=" + ctx.Value(key1).(string))
		for {
			select {
			case <-ctx.Done():
				fmt.Println("canncelled")
				done <- struct{}{}
				return
			default:
			}
			v := ctx.Value(key2).(*ctxData)
			v.value += "."
			fmt.Println(v.value)
			time.Sleep(time.Duration(300) * time.Millisecond)
		}
	}()

	<-done
	fmt.Println("context value test done.")
}

/*
time.ticker
*/

func demo0501() {
	ticker := time.NewTicker(time.Second)
	i := 0
	for range ticker.C {
		fmt.Println("do sometime each second")
		i++
		if i > 3 {
			ticker.Stop()
			return
		}
	}
}

func demo0502() {
	timeout := 5
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	go func(ctx context.Context) {
		c := time.Tick(time.Second)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("routine exit")
				return
			case <-c:
				fmt.Println("routine sleep for second")
			}
		}
	}(ctx)

	time.Sleep(time.Duration(timeout+1) * time.Second)
	fmt.Println("Done")
}

/*
rpc by context
*/

func demo06() {
	start := time.Now()
	timeout := 8
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	rpcCalls(ctx, cancel)

	select {
	case <-ctx.Done():
		fmt.Println("duration:", time.Now().Sub(start).Seconds())
		fmt.Println("rpc calls done")
	case <-time.After(time.Duration(timeout-1) * time.Second):
		fmt.Println("timeout")
		cancel()
	}
}

func rpcCalls(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mockRPC(ctx, "http://zj.test1.com"); err != nil {
			fmt.Println("rpc call failed, error:", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := mockRPC(ctx, "http://zj.test2.com"); err != nil {
			fmt.Println("rpc call failed, error:", err)
			cancel()
		}
	}()
	wg.Wait()
}

func mockRPC(ctx context.Context, url string) error {
	fmt.Println("mock rpc call for:", url)
	result := make(chan int)
	err := make(chan error)

	go func() {
		time.Sleep(time.Second)
		isSuccess := true
		if isSuccess {
			result <- 1
		} else {
			err <- fmt.Errorf("mock error")
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e := <-err:
		return e
	case <-result:
		return nil
	}
}

/*
channel close

1. channel不能close两次，否则panic
2. 读取的时候channel提前关闭了，不会panic. 返回值：string => "", bool => false
3. 向已经关闭的channel写数据，会引起panic
4. 判断channel是否close: if value, ok := <- ch; !ok { fmt.Println("closed") }
5. for循环读取channel, ch关闭时，for循环会自动结束
*/

func demo0701() {
	c := make(chan int)
	go func(c chan<- int) {
		defer close(c)
		for i := 0; i < 3; i++ {
			time.Sleep(time.Duration(300) * time.Millisecond)
			fmt.Println("send:", i)
			c <- i
		}
	}(c)

	for v := range c {
		fmt.Println("receive:", v)
	}
}

func demo0702() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("error:", err)
		}
	}()

	c := make(chan int)
	go func(c <-chan int) {
		for n := range c {
			if n > 3 {
				return
			}
			fmt.Println("receive:", n)
		}
	}(c)

	go func() {
		time.Sleep(time.Duration(2) * time.Second)
		fmt.Println("timeout and close channel")
		close(c)
	}()

	for i := 0; i < 10; i++ {
		fmt.Println("send:", i)
		c <- i
	}
}

// demo08, handle panic in goroutine
func demo08() {
	// 1. if goroutine panic, stack will print, and main process will exit
	// 2. we can only handle panic in current goroutine
	myPrint := func(num int) {
		if num%7 == 0 {
			panic("sub mock exception")
		}
		fmt.Println(num)
	}

	handler := func(num int) {
		if num%5 == 0 {
			panic("mock exception")
		}
		go myPrint(num)
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("catch exception:", err)
				stackTrace := debug.Stack() // get stack
				fmt.Fprintln(os.Stderr, string(stackTrace))
			}
		}()
		nums := []int{12, 2, 41, 5, 7, 6, 1, 12, 3, 5, 2, 31}
		for _, num := range nums {
			fmt.Println("current number:", num)
			handler(num)
		}
	}()
	time.Sleep(time.Duration(2) * time.Second)
	fmt.Println("demo 08 finished")
}

// demo09, error with stack
func demo09() {
	count := 2
	rand.Seed(time.Now().Unix())
	errChan := make(chan error, count)

	wg := sync.WaitGroup{}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			n := rand.Intn(3)
			time.Sleep(time.Duration(n) * time.Second)
			err := fmt.Errorf("mock error from: index=%d", idx)
			errChan <- errors.WithStack(err)
		}(i)
	}
	wg.Wait()
	close(errChan)

	for i := 0; i < count+2; i++ {
		if err, ok := <-errChan; ok {
			// fmt.Printf("error: %v\n", errors.Cause(err))
			fmt.Printf("error: %v\n", err)
			fmt.Printf("stack: %+v\n", err)
			fmt.Println()
		} else {
			fmt.Println("channel closed")
			break
		}
	}
}

/*
func deco
*/

func deco(fn func(string) string) func(string) string {
	return func(text string) string {
		fmt.Println("pre-hook")
		ret := fn(text)
		fmt.Println("after-hook")
		return ret
	}
}

func demo10() {
	buildHello := func(name string) string {
		fmt.Println("build hello message...")
		return "Hello " + name
	}
	decoFunc := deco(buildHello)
	fmt.Println(decoFunc("Henry"))
}

// demo11, atomic int
func demo11() {
	var v int32
	atomic.StoreInt32(&v, 10)

	parallel := 3
	wg := sync.WaitGroup{}
	wg.Add(parallel)
	for i := 0; i < parallel; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				atomic.AddInt32(&v, 1)
			}
		}()
	}
	wg.Wait()

	res := atomic.LoadInt32(&v)
	if atomic.CompareAndSwapInt32(&v, res, res+1) {
		fmt.Println("atomic value:", atomic.LoadInt32(&v))
	} else {
		fmt.Println("atomic CompareAndSwapInt32 failed")
	}
}

// demo12, struct func split in 2 files
type myPerson struct {
	Name   string
	Age    int
	Skills []string
}

func (p myPerson) SayHello() {
	p.Name = strings.Title(p.Name)
	fmt.Println("hello, my name is:", p.Name)
}

// Main .
func Main() {
	fmt.Println("demo main")
	demo09()
}
