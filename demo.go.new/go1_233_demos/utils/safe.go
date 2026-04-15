package utils

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// SafeCounter is a thread-safe counter impl by CAS.

type SafeCounter struct {
	value int32
}

func (c *SafeCounter) Increment() {
	for {
		old := atomic.LoadInt32(&c.value)
		val := old + 1
		if atomic.CompareAndSwapInt32(&c.value, old, val) {
			return
		}
		runtime.Gosched()
	}
}

func (c *SafeCounter) Value() int32 {
	return atomic.LoadInt32(&c.value)
}

// SpinLock is a simple spin lock impl by CAS.

const (
	SpinUnlocked = iota
	SpinLocked
)

type SpinLock struct {
	flag int32
}

func (l *SpinLock) Lock() {
	for !atomic.CompareAndSwapInt32(&l.flag, SpinUnlocked, SpinLocked) {
		runtime.Gosched()
	}
}

func (l *SpinLock) Unlock() {
	atomic.StoreInt32(&l.flag, SpinUnlocked)
}

// Generic Sync Pool
//
// 在泛型池中建议存储指针类型 (如 *MyStruct), 因为 sync.Pool 存储值类型 (Value Type) 时仍会触发逃逸分析到堆上的分配, 无法完全规避 GC.

type SyncPool[T any] struct {
	internal sync.Pool
}

func NewSyncPool[T any](alloc func() T) *SyncPool[T] {
	return &SyncPool[T]{
		internal: sync.Pool{
			New: func() any {
				return alloc()
			},
		},
	}
}

func (p *SyncPool[T]) Get() T {
	val, _ := p.internal.Get().(T)
	return val
}

func (p *SyncPool[T]) Put(x T, reset func(T)) {
	if reset != nil {
		reset(x)
	}
	p.internal.Put(x)
}
