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

// GoRoutinePool .
type GoRoutinePool interface {
	Submit(context.Context, Function, ...interface{}) (chan interface{}, error)
	Start()
	Stop()
	Cancel()
	Usage() string
}

// PrintPoolInfo .
func PrintPoolInfo(pool GoRoutinePool) {
	fmt.Println("usage:", pool.Usage())
}

/*
Pool with Semaphore:
1. always start a new goroutine to run submitted func.
2. for numbers of goroutines which are exceed coresize, they are blocked.
3. if number of submitted funcs exceeds maxsize, they will be discard.
*/

// Function .
type Function func(context.Context, ...interface{}) interface{}

// GoRoutinePoolWithSemaphore .
type GoRoutinePoolWithSemaphore struct {
	semaphore    chan struct{}
	globalCtx    context.Context
	cancelFunc   context.CancelFunc
	wg           sync.WaitGroup
	isRunning    bool
	maxSize      int
	numOfWaiting int32
}

// NewGoRoutinePoolWithSemaphore creates goroutine pool by coreSize (parallel number) and maxSize (max number of submitted funcs).
func NewGoRoutinePoolWithSemaphore(coreSize, maxSize int) *GoRoutinePoolWithSemaphore {
	ctx, cancel := context.WithCancel(context.Background())
	return &GoRoutinePoolWithSemaphore{
		semaphore:  make(chan struct{}, coreSize),
		globalCtx:  ctx,
		cancelFunc: cancel,
		wg:         sync.WaitGroup{},
		maxSize:    maxSize,
	}
}

// Submit submits a func into pool to process.
func (pool *GoRoutinePoolWithSemaphore) Submit(ctx context.Context, fn Function, args ...interface{}) (chan interface{}, error) {
	if !pool.isRunning {
		return nil, errors.New("pool is not started")
	}

	ch := make(chan interface{}, 1)
	num := atomic.AddInt32(&pool.numOfWaiting, 1)
	// 防止启动太多的 goroutines 占用 fh
	if int(num) > pool.maxSize-len(pool.semaphore) {
		atomic.AddInt32(&pool.numOfWaiting, -1)
		ch <- fmt.Sprintf("exceed max size %d, and discard", pool.maxSize)
		return ch, nil
	}

	// 根据 submit 的任务数量启动对应数量的 goroutine 执行任务
	// 当超过 semaphore 时，goroutine 处于阻塞状态
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
func (pool *GoRoutinePoolWithSemaphore) Start() {
	pool.isRunning = true
}

// Stop waits all running funcs done, and exit goroutine pool.
func (pool *GoRoutinePoolWithSemaphore) Stop() {
	pool.isRunning = false
	pool.wait()
}

// Cancel cancels all submitted funcs which are pending to process, and stop.
func (pool *GoRoutinePoolWithSemaphore) Cancel() {
	pool.isRunning = false
	pool.cancelFunc()
	pool.wait()
}

// Usage .
func (pool *GoRoutinePoolWithSemaphore) Usage() string {
	return fmt.Sprintf("wait/run/idle:%d/%d/%d\n",
		pool.numOfWaiting, len(pool.semaphore), cap(pool.semaphore)-len(pool.semaphore))
}

func (pool *GoRoutinePoolWithSemaphore) wait() {
	pool.wg.Wait()
}

/*
Pool with fix size:
1. start fixed core size of goroutines to process func.
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

// GoRoutinePoolWithFixSize .
type GoRoutinePoolWithFixSize struct {
	coreSize     int
	queue        chan FunctionEvent
	isRunning    bool
	isCancelled  bool
	numOfRunning int32
}

// NewGoRoutinePoolWithFixSize creates goroutine pool by coreSize (parallel number) and queueSize (stores waitting funcs).
func NewGoRoutinePoolWithFixSize(coreSize, queueSize int) *GoRoutinePoolWithFixSize {
	return &GoRoutinePoolWithFixSize{
		coreSize: coreSize,
		queue:    make(chan FunctionEvent, queueSize),
	}
}

// Submit .
func (pool *GoRoutinePoolWithFixSize) Submit(ctx context.Context, fn Function, args ...interface{}) (chan interface{}, error) {
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
func (pool *GoRoutinePoolWithFixSize) Start() {
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
func (pool *GoRoutinePoolWithFixSize) Stop() {
	pool.isRunning = false
	pool.wait()
	close(pool.queue)
}

// Cancel waits all running funcs done, then cancels funcs in queue, and stop.
func (pool *GoRoutinePoolWithFixSize) Cancel() {
	pool.isCancelled = true
	pool.Stop()
}

// Usage .
func (pool *GoRoutinePoolWithFixSize) Usage() string {
	return fmt.Sprintf("wait/run/idle:%d/%d/%d\n",
		len(pool.queue), pool.numOfRunning, pool.coreSize-int(pool.numOfRunning))
}

func (pool *GoRoutinePoolWithFixSize) wait() {
	for pool.numOfRunning != 0 {
		time.Sleep(time.Second)
	}
}

/*
Go Pool:
1. when submit a task
  - if no running worker, start a new goroutine
  - if there are running workers, reuse them
2. number of goroutines is controlled by semaphore
3. if no more tasks, goroutines will be exit
*/

// Worker .
type Worker struct {
	work      chan func()
	semaphore chan struct{}
}

// GoPool .
type GoPool struct {
	Worker
	isRunning bool
	idleTime  time.Duration
	queue     chan func()
	stopCh    chan struct{}
}

// NewGoPool .
func NewGoPool(coreSize, maxSize int, idleTime time.Duration) *GoPool {
	pool := &GoPool{
		idleTime: idleTime,
		queue:    make(chan func(), maxSize-coreSize),
		stopCh:   make(chan struct{}),
		Worker: Worker{
			work:      make(chan func()),
			semaphore: make(chan struct{}, coreSize),
		},
	}
	pool.Start()
	return pool
}

// Submit .
func (pool *GoPool) Submit(task func()) error {
	if !pool.isRunning {
		return fmt.Errorf("pool is not running")
	}

	select {
	case pool.queue <- task:
	default:
		return fmt.Errorf("exceed max size %d, and discard", cap(pool.semaphore)+cap(pool.queue))
	}
	return nil
}

// SubmitWithTimeout .
func (pool *GoPool) SubmitWithTimeout(task func(), timeout time.Duration) error {
	if !pool.isRunning {
		return fmt.Errorf("pool is not running")
	}

	select {
	case pool.queue <- task:
	case <-time.After(timeout):
		return fmt.Errorf("timeout: exceed max size %d, and discard", cap(pool.semaphore)+cap(pool.queue))
	}
	return nil
}

// Start .
func (pool *GoPool) Start() {
	pool.isRunning = true
	go pool.run()
}

func (pool *GoPool) run() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("[GoPool]: run error:", err)
		}
	}()

	for task := range pool.queue {
		select {
		// if ch "work" or "semaphore" is closed when get task, panic here
		case pool.work <- task:
		case pool.semaphore <- struct{}{}:
			go pool.worker(task)
		}
	}
}

func (pool *GoPool) worker(task func()) {
	workerID := time.Now().Nanosecond()
	fmt.Printf("[worker %d]: init\n", workerID)
	defer func() {
		<-pool.semaphore
	}()

	tick := time.NewTicker(pool.idleTime)
	defer tick.Stop()

	localTask := task
	for {
		localTask()
		select {
		case <-pool.stopCh:
			fmt.Printf("[worker %d]: stop\n", workerID)
			return
		case <-tick.C:
			// if there is no more tasks for N sec, then worker exit
			fmt.Printf("[worker %d]: idle and exit\n", workerID)
			return
		case localTask = <-pool.work:
			fmt.Printf("[worker %d]: fetch task\n", workerID)
			tick.Reset(pool.idleTime)
		}
	}
}

// Stop waits for seconds and stops go pool.
// 1. not allow submit more tasks
// 2. wait seconds for exitsing tasks done
// 3. close chan and exit
// Note: task itself should make sure it can be cancelled by "ctx", or it will be leak after GoPool stop.
func (pool *GoPool) Stop(waitSec int) {
	pool.isRunning = false
	close(pool.stopCh)
	for i := 0; i < waitSec; i++ {
		if len(pool.semaphore) == 0 {
			break
		}
		time.Sleep(time.Second)
	}
	close(pool.queue)
	close(pool.work)
	close(pool.semaphore)
}

// Cancel .
func (pool *GoPool) Cancel() {
	pool.Stop(3)
}

// Usage .
func (pool *GoPool) Usage() string {
	return fmt.Sprintf("wait/run/idle:%d/%d/0", len(pool.queue), len(pool.semaphore))
}
