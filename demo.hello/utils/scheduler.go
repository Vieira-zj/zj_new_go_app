package utils

import (
	"time"
)

// Task .
type Task func()

// Scheduler run scheduled tasks by fix interval. (执行定时任务)
type Scheduler struct {
	timers map[string]*time.Timer
}

// NewScheduler .
func NewScheduler() *Scheduler {
	return &Scheduler{
		timers: make(map[string]*time.Timer, 16),
	}
}

// AddTask adds and runs task at fix interval.
func (s *Scheduler) AddTask(interval time.Duration, name string, task Task) {
	s.timers[name] = time.AfterFunc(interval, func() {
		task()
		s.timers[name].Reset(interval)
	})
}

// StopTask stop schedule task by name.
func (s *Scheduler) StopTask(name string) {
	s.timers[name].Stop()
}

// Close .
func (s *Scheduler) Close() {
	for _, timer := range s.timers {
		timer.Stop()
	}
}
