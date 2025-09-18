package utils

import (
	"runtime"
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

type SpinLock struct {
	flag int32
}

func (l *SpinLock) Lock() {
	for !atomic.CompareAndSwapInt32(&l.flag, 0, 1) {
		runtime.Gosched()
	}
}

func (l *SpinLock) Unlock() {
	atomic.StoreInt32(&l.flag, 0)
}
