package utils

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

/*
Pool interface
*/

// RoutinePool .
type RoutinePool interface {
	Submit(context.Context, Function, ...interface{}) (chan interface{}, error)
	Start()
	Stop()
	Cancel()
	Usage() string
}

// PrintPoolInfo .
func PrintPoolInfo(pool RoutinePool) {
	fmt.Println("usage:", pool.Usage())
}

/*
Pool with Semaphore:
1. always start a new routinue to run submitted func.
2. for numbers of routinues which are exceed coresize, they are blocked.
3. if number of submitted funcs exceeds maxsize, they will be discard.
*/

// Function .
type Function func(context.Context, ...interface{}) interface{}

// RoutinuePoolWithSemaphore .
type RoutinuePoolWithSemaphore struct {
	semaphore    chan struct{}
	globalCtx    context.Context
	cancelFunc   context.CancelFunc
	wg           sync.WaitGroup
	isRunning    bool
	maxSize      int
	numOfWaiting int32
}

// NewRoutinuePoolWithSemaphore creates routinue pool by coreSize (parallel number) and maxSize (max number of submitted funcs).
func NewRoutinuePoolWithSemaphore(coreSize, maxSize int) *RoutinuePoolWithSemaphore {
	ctx, cancel := context.WithCancel(context.Background())
	return &RoutinuePoolWithSemaphore{
		semaphore:  make(chan struct{}, coreSize),
		globalCtx:  ctx,
		cancelFunc: cancel,
		wg:         sync.WaitGroup{},
		maxSize:    maxSize,
	}
}

// Submit submits a func into pool to process.
func (pool *RoutinuePoolWithSemaphore) Submit(ctx context.Context, fn Function, args ...interface{}) (chan interface{}, error) {
	if !pool.isRunning {
		return nil, errors.New("pool is not started")
	}

	ch := make(chan interface{}, 1)
	num := atomic.AddInt32(&pool.numOfWaiting, 1)
	// 防止启动太多的 routinues 占用 fh
	if int(num) > pool.maxSize-len(pool.semaphore) {
		atomic.AddInt32(&pool.numOfWaiting, -1)
		ch <- fmt.Sprintf("exceed max size %d, and discard", pool.maxSize)
		return ch, nil
	}

	// 根据 submit 的任务数量启动对应数量的 routinue 执行任务
	// 当超过 semaphore 时，routinue 处于阻塞状态
	pool.wg.Add(1)
	go func() {
		defer pool.wg.Done()
		select {
		case pool.semaphore <- struct{}{}:
			defer func() {
				<-pool.semaphore
			}()
			atomic.AddInt32(&pool.numOfWaiting, -1)
			ch <- fn(ctx, args...)
		case <-ctx.Done():
			atomic.AddInt32(&pool.numOfWaiting, -1)
			ch <- "timeout"
		case <-pool.globalCtx.Done():
			atomic.AddInt32(&pool.numOfWaiting, -1)
			ch <- "cancelled"
		}
	}()
	return ch, nil
}

// Start .
func (pool *RoutinuePoolWithSemaphore) Start() {
	pool.isRunning = true
}

// Stop waits all running funcs done, and exit routinue pool.
func (pool *RoutinuePoolWithSemaphore) Stop() {
	pool.isRunning = false
	pool.wait()
}

// Cancel cancels all submitted funcs which are pending to process, and stop.
func (pool *RoutinuePoolWithSemaphore) Cancel() {
	pool.isRunning = false
	pool.cancelFunc()
	pool.wait()
}

// Usage .
func (pool *RoutinuePoolWithSemaphore) Usage() string {
	return fmt.Sprintf("wait/run/idle:%d/%d/%d\n",
		pool.numOfWaiting, len(pool.semaphore), cap(pool.semaphore)-len(pool.semaphore))
}

func (pool *RoutinuePoolWithSemaphore) wait() {
	pool.wg.Wait()
}

/*
Pool with fix size:
1. start fixed core size of routinues to process func.
2. for submitted funcs which exceed coresize, put them into queue. (no blocked)
3. if number of submitted funcs exceeds maxsize, they will be discard.
*/

// FunctionEvent .
type FunctionEvent struct {
	Ctx  context.Context
	Fn   Function
	Args []interface{}
	Ret  chan interface{}
}

// RoutinuePoolWithFixSize .
type RoutinuePoolWithFixSize struct {
	coreSize     int
	queue        chan FunctionEvent
	isRunning    bool
	isCancelled  bool
	numOfRunning int32
}

// NewRoutinuePoolWithFixSize creates routinue pool by coreSize (parallel number) and queueSize (stores waitting funcs).
func NewRoutinuePoolWithFixSize(coreSize, queueSize int) *RoutinuePoolWithFixSize {
	return &RoutinuePoolWithFixSize{
		coreSize: coreSize,
		queue:    make(chan FunctionEvent, queueSize),
	}
}

// Submit .
func (pool *RoutinuePoolWithFixSize) Submit(ctx context.Context, fn Function, args ...interface{}) (chan interface{}, error) {
	if !pool.isRunning {
		return nil, errors.New("pool is not started")
	}

	ret := make(chan interface{}, 1)
	time.Sleep(time.Duration(10) * time.Millisecond) // wait for previous submit done
	if len(pool.queue) == cap(pool.queue) {
		ret <- fmt.Sprintf("exceed max size %d, and discard", pool.coreSize+cap(pool.queue))
		return ret, nil
	}

	go func() {
		event := FunctionEvent{
			Ctx:  ctx,
			Fn:   fn,
			Args: args,
			Ret:  ret,
		}
		pool.queue <- event
	}()
	return ret, nil
}

// Start .
func (pool *RoutinuePoolWithFixSize) Start() {
	pool.isRunning = true
	for i := 0; i < pool.coreSize; i++ {
		go func() {
			for event := range pool.queue { // loop
				if pool.isCancelled {
					event.Ret <- "cancelled"
					continue
				}
				atomic.AddInt32(&pool.numOfRunning, 1)
				event.Ret <- event.Fn(event.Ctx, event.Args...)
				atomic.AddInt32(&pool.numOfRunning, -1)
			}
		}()
	}
}

// Stop waits all funcs (run + queue) done, and exit.
func (pool *RoutinuePoolWithFixSize) Stop() {
	pool.isRunning = false
	pool.wait()
	close(pool.queue)
}

// Cancel waits all running funcs done, then cancels funcs in queue, and stop.
func (pool *RoutinuePoolWithFixSize) Cancel() {
	pool.isCancelled = true
	pool.Stop()
}

// Usage .
func (pool *RoutinuePoolWithFixSize) Usage() string {
	return fmt.Sprintf("wait/run/idle:%d/%d/%d\n",
		len(pool.queue), pool.numOfRunning, pool.coreSize-int(pool.numOfRunning))
}

func (pool *RoutinuePoolWithFixSize) wait() {
	for pool.numOfRunning != 0 {
		time.Sleep(time.Second)
	}
}
