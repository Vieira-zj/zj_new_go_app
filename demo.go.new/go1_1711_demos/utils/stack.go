package utils

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

// SyncStack sync stack demo by sync Condition.
type SyncStack struct {
	elements     []interface{}
	notEmptyCond *sync.Cond // 队列未空条件队列
	notFullCond  *sync.Cond // 队列未满条件队列
}

func NewSyncStack(capability int) *SyncStack {
	lock := sync.Mutex{}
	notEmptyCond := sync.NewCond(&lock)
	notFullCond := sync.NewCond(&lock)

	return &SyncStack{
		elements:     make([]interface{}, 0, capability),
		notEmptyCond: notEmptyCond,
		notFullCond:  notFullCond,
	}
}

func (s *SyncStack) Size() int {
	return len(s.elements)
}

func (s *SyncStack) isFull() bool {
	return s.Size() >= cap(s.elements)
}

func (s *SyncStack) isEmpty() bool {
	return s.Size() == 0
}

func (s *SyncStack) String() string {
	list := make([]string, 0, s.Size())
	for _, ele := range s.elements {
		list = append(list, fmt.Sprintf("%v", ele))
	}
	return strings.Join(list, ",")
}

func (s *SyncStack) Close() {
	// 释放 cond wait 的 goroutine
	s.notEmptyCond.L.Lock()
	defer s.notEmptyCond.L.Unlock()

	s.notEmptyCond.Broadcast()
	s.notFullCond.Broadcast()
}

func (s *SyncStack) Pop(ctx context.Context) (ele interface{}, err error) {
	s.notEmptyCond.L.Lock()
	defer func() {
		// Wait 内部已经释放了锁，避免 unlock of unlocked mutex 错误
		if originErr := errors.Cause(err); isContextCancel(originErr) {
			return
		}
		s.notEmptyCond.L.Unlock()
	}()

	for s.isEmpty() {
		// 如果队列是空的，就在 notEmptyCond 条件上等待
		// Wait 内部会先释放锁，等到收到满足信号时将重新尝试获得锁
		if err = condWaitWithCancel(ctx, s.notEmptyCond); err != nil {
			return
		}
	}

	ele = s.elements[s.Size()-1]
	s.elements = s.elements[:s.Size()-1]
	// 此时队列中已经 Pop 一个值，不再满，发送 notFullCond 信号激活再此条件 Wait 的操作
	s.notFullCond.Signal()
	return
}

func (s *SyncStack) Add(ctx context.Context, ele interface{}) (err error) {
	s.notEmptyCond.L.Lock()
	defer func() {
		// Wait 内部已经释放了锁，避免 unlock of unlocked mutex 错误
		if originalErr := errors.Cause(err); isContextCancel(originalErr) {
			return
		}
		s.notEmptyCond.L.Unlock()
	}()

	for s.isFull() {
		if err = condWaitWithCancel(ctx, s.notFullCond); err != nil {
			return
		}
	}

	s.elements = append(s.elements, ele)
	// 此时队列中已经有值，发送队列不为空的信号激活再此条件 Wait 的操作
	s.notEmptyCond.Signal()
	return
}

func isContextCancel(err error) bool {
	return err == context.DeadlineExceeded || err == context.Canceled
}

func condWaitWithCancel(ctx context.Context, cond *sync.Cond) error {
	if ctx.Done() != nil {
		waitDone := make(chan struct{})
		go func() {
			// ctx cancel 后，这里 condition 一直 wait, goroutine 会泄漏
			// 执行 stack.Close() 释放
			cond.Wait()
			close(waitDone)
			cond.L.Unlock() // 注：wait 返回后，这里要释放锁
			log.Println("exit cond wait")
		}()

		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "cancel wait")
		case <-waitDone:
			return nil
		}
	}

	cond.Wait()
	return nil
}
