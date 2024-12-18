package demos

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
)

// Broadcaster by sync.Cond

type Broadcaster1 struct {
	mu       *sync.Mutex
	cond     *sync.Cond
	signaled bool
}

func NewBroadcaster1() *Broadcaster1 {
	var mu sync.Mutex
	return &Broadcaster1{
		mu:       &mu,
		cond:     sync.NewCond(&mu),
		signaled: false,
	}
}

func (b *Broadcaster1) Go(fn func()) {
	if b.signaled {
		log.Println("broadcaster already exit")
		return
	}

	go func() {
		b.cond.L.Lock()
		defer b.cond.L.Unlock()

		for !b.signaled {
			log.Println("wait...")
			b.cond.Wait()
		}
		fn()
	}()
}

func (b *Broadcaster1) Broadcast() {
	if b.signaled {
		log.Println("broadcaster already exit")
		return
	}

	b.cond.L.Lock()
	b.signaled = true
	b.cond.L.Unlock()

	b.cond.Broadcast()
}

// Broadcaster by Channel

type Broadcaster2 struct {
	signal chan struct{}
}

func NewBroadcaster2() *Broadcaster2 {
	return &Broadcaster2{
		signal: make(chan struct{}),
	}
}

func (b *Broadcaster2) Go(fn func()) {
	go func() {
		log.Println("wait...")
		<-b.signal
		fn()
	}()
}

func (b *Broadcaster2) Broadcast() {
	close(b.signal)
}

// Broadcaster by Context

type Broadcaster3 struct {
	key    struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

func NewBroadcaster3() *Broadcaster3 {
	ctx, cancel := context.WithCancel(context.Background())
	return &Broadcaster3{
		key:    struct{}{},
		ctx:    ctx,
		cancel: cancel,
	}
}

func (b *Broadcaster3) Go(fn func()) {
	if val, ok := b.ctx.Value(b.key).(bool); ok && val {
		log.Println("broadcaster already exit")
		return
	}

	go func() {
		log.Println("wait...")
		<-b.ctx.Done()
		fn()
	}()
}

func (b *Broadcaster3) Broadcast() {
	if val, ok := b.ctx.Value(b.key).(bool); ok && val {
		log.Println("broadcaster already exit")
		return
	}

	b.ctx = context.WithValue(b.ctx, b.key, true)
	b.cancel()
}

// Broadcaster by wait.Group

type Broadcaster4 struct {
	done    uint32
	trigger sync.WaitGroup
}

func NewBroadcaster4() *Broadcaster4 {
	b := &Broadcaster4{}
	b.trigger.Add(1)
	return b
}

func (b *Broadcaster4) Go(fn func()) {
	go func() {
		if atomic.LoadUint32(&b.done) == 1 {
			log.Println("broadcaster already exit")
			return
		}

		b.trigger.Wait()
		fn()
	}()
}

func (b *Broadcaster4) Broadcast() {
	if atomic.LoadUint32(&b.done) == 1 {
		log.Println("broadcaster already exit")
		return
	}

	if atomic.CompareAndSwapUint32(&b.done, 0, 1) {
		b.trigger.Done()
	}
}

// Broadcaster by sync.RWMutex

type Broadcaster5 struct {
	mu *sync.RWMutex
}

func NewBroadcaster5() *Broadcaster5 {
	var mu sync.RWMutex
	mu.Lock()

	return &Broadcaster5{mu: &mu}
}

func (b *Broadcaster5) Go(fn func()) {
	go func() {
		b.mu.RLock()
		defer b.mu.RUnlock()
		fn()
	}()
}

func (b *Broadcaster5) Broadcast() {
	b.mu.Unlock()
}
