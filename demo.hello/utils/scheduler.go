package utils

import (
	"context"
	"fmt"
	"time"
)

// Task .
type Task func()

// Scheduler .
type Scheduler interface {
	AddTask(string, time.Duration, Task)
	StopTask(string)
	Close()
}

//
// Scheduler by time.AfterFunc
//

// SchedulerV1 run scheduled tasks by fix interval.
type SchedulerV1 struct {
	timers map[string]*time.Timer
}

// NewSchedulerV1 .
func NewSchedulerV1() *SchedulerV1 {
	return &SchedulerV1{
		timers: make(map[string]*time.Timer, 8),
	}
}

// AddTask adds and runs task at fix interval.
func (s *SchedulerV1) AddTask(name string, interval time.Duration, task Task) {
	s.timers[name] = time.AfterFunc(interval, func() {
		task()
		if t, ok := s.timers[name]; ok {
			t.Reset(interval)
		} else {
			panic(fmt.Sprintf("task [%s] is nil\n", name))
		}
	})
}

// StopTask stop schedule task by name.
func (s *SchedulerV1) StopTask(name string) {
	if t, ok := s.timers[name]; ok {
		t.Stop()
	}
}

// Close .
func (s *SchedulerV1) Close() {
	for _, timer := range s.timers {
		timer.Stop()
	}
}

//
// Scheduler by time.Tick
//

// NewSchedulerV2 .
func NewSchedulerV2() *SchedulerV2 {
	return &SchedulerV2{
		cannels: make(map[string]context.CancelFunc, 16),
	}
}

// SchedulerV2 .
type SchedulerV2 struct {
	cannels map[string]context.CancelFunc
}

// AddTask .
func (s *SchedulerV2) AddTask(name string, interval time.Duration, task Task) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		tick := time.Tick(interval)
		for {
			select {
			case <-tick:
				task()
			case <-ctx.Done():
				fmt.Printf("task [%s] exit\n", name)
				return
			}
		}
	}()
	s.cannels[name] = cancel
}

// StopTask .
func (s *SchedulerV2) StopTask(name string) {
	if cancel, ok := s.cannels[name]; ok {
		cancel()
	}
}

// Close .
func (s *SchedulerV2) Close() {
	for _, cancel := range s.cannels {
		cancel()
	}
}
