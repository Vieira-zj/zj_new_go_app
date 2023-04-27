package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	PollStatusPending = "pending"
	PollStatusReady   = "ready"
)

// Waker

type Waker struct {
	queue chan Future
}

func (w Waker) wake(f Future) {
	w.queue <- f
	// fmt.Println("Waker: waked")
}

// Future

type Future interface {
	poll() string
	string() string
}

type TimerFuture struct {
	index     int
	runIdx    int
	completed bool
	waker     Waker
}

func NewTimerFuture(idx int, queue chan Future) *TimerFuture {
	return &TimerFuture{
		index: idx,
		waker: Waker{
			queue: queue,
		},
	}
}

func (f *TimerFuture) string() string {
	return fmt.Sprintf("future-%d", f.index)
}

func (f *TimerFuture) poll() string {
	if f.completed {
		return PollStatusReady
	}
	go f.run() // 模拟 io 阻塞处理
	return PollStatusPending
}

func (f *TimerFuture) run() {
	f.runIdx += 1
	fmt.Printf("[%s.thread] processing...\n", f.string())
	time.Sleep(getSleep() * time.Second)

	if f.runIdx > 2 {
		fmt.Printf("[%s.thread] completed and wake\n", f.string())
		f.completed = true
		f.waker.wake(f)
	} else {
		fmt.Printf("[%s.thread] wake\n", f.string())
		f.waker.wake(f)
	}
}

func getSleep() time.Duration {
	sleep := rand.Intn(3)
	if sleep == 0 {
		return 1
	}
	return time.Duration(sleep)
}
