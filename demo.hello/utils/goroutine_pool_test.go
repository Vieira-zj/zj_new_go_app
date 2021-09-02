package utils

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func process(ctx context.Context, sec int) string {
	select {
	case <-ctx.Done():
		return "timeout"
	default:
		text := fmt.Sprintf("sec=%d", sec)
		fmt.Println("process start:", text)
		time.Sleep(time.Duration(sec) * time.Second)
		return text
	}
}

func buildFuncFromProcess(p func(context.Context, int) string) Function {
	return func(ctx context.Context, args ...interface{}) interface{} {
		sec := (args[0]).(int)
		return p(ctx, sec)
	}
}

func TestGoRoutinePoolWithSemaphore(t *testing.T) {
	pool := NewGoRoutinePoolWithSemaphore(2, 10)
	testGoRoutinePool(t, pool)
}

func TestGoRoutinePoolWithSemaphoreByDiscard(t *testing.T) {
	pool := NewGoRoutinePoolWithSemaphore(2, 5)
	testGoRoutinePool(t, pool)
}

func TestGoRoutinePoolWithFixSize(t *testing.T) {
	pool := NewGoRoutinePoolWithFixSize(2, 10)
	testGoRoutinePool(t, pool)
}

func TestGoRoutinePoolWithFixSizeByDiscard(t *testing.T) {
	pool := NewGoRoutinePoolWithFixSize(2, 3)
	testGoRoutinePool(t, pool)
}

func testGoRoutinePool(t *testing.T, pool GoRoutinePool) {
	count := 6
	rand.Seed(time.Now().Unix())
	retChList := make([]chan interface{}, 0, count)
	fn := buildFuncFromProcess(process)

	pool.Start()
	defer pool.Stop()
	for i := 0; i < count; i++ {
		num := rand.Int31n(5)
		if num == 0 {
			num = 1
		}
		ret, err := pool.Submit(context.Background(), fn, int(num))
		if err != nil {
			t.Fatal(err)
		}
		retChList = append(retChList, ret)
	}

	wg := sync.WaitGroup{}
	for _, ch := range retChList {
		wg.Add(1)
		go func(ch chan interface{}) {
			defer wg.Done()
			fmt.Println("results:", <-ch)
		}(ch)
	}

	for i := 0; i < 10; i++ {
		PrintPoolInfo(pool)
		time.Sleep(time.Second)
	}
	wg.Wait()
}

func TestGoRoutinePoolWithSemaphoreByTimeout(t *testing.T) {
	pool := NewGoRoutinePoolWithSemaphore(2, 10)
	testGoRoutinePoolWithTimeout(t, pool)
}

func TestGoRoutinePoolWithFixSizeByTimeout(t *testing.T) {
	pool := NewGoRoutinePoolWithFixSize(2, 10)
	testGoRoutinePoolWithTimeout(t, pool)
}

func testGoRoutinePoolWithTimeout(t *testing.T, pool GoRoutinePool) {
	count := 6
	rand.Seed(time.Now().Unix())
	retChList := make([]chan interface{}, 0, count)
	fn := buildFuncFromProcess(process)

	pool.Start()
	defer pool.Stop()
	for i := 0; i < count; i++ {
		num := rand.Int31n(5)
		if num == 0 {
			num = 1
		}
		// 超过 3s 没有被执行的任务会被 cancel 掉
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
		defer cancel()
		ret, err := pool.Submit(ctx, fn, int(num))
		if err != nil {
			t.Fatal(err)
		}
		retChList = append(retChList, ret)
	}

	for _, ch := range retChList {
		fmt.Println("results:", <-ch)
	}
}

func TestGoRoutinePoolWithSemaphoreByCancel(t *testing.T) {
	pool := NewGoRoutinePoolWithSemaphore(2, 10)
	testGoRoutinePoolWithCancel(t, pool)
}

func TestGoRoutinePoolWithFixSizeByCancel(t *testing.T) {
	pool := NewGoRoutinePoolWithFixSize(2, 10)
	testGoRoutinePoolWithCancel(t, pool)
}

func testGoRoutinePoolWithCancel(t *testing.T, pool GoRoutinePool) {
	count := 6
	rand.Seed(time.Now().Unix())
	retChList := make([]chan interface{}, 0, count)
	fn := buildFuncFromProcess(process)
	pool.Start()

	for i := 0; i < count; i++ {
		num := rand.Int31n(5)
		if num == 0 {
			num = 1
		}
		ret, err := pool.Submit(context.Background(), fn, int(num))
		if err != nil {
			t.Fatal(err)
		}
		retChList = append(retChList, ret)
	}

	time.Sleep(time.Duration(2) * time.Second)
	pool.Cancel()

	for _, ch := range retChList {
		fmt.Println("results:", <-ch)
	}
}
