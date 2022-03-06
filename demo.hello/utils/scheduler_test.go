package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestRunSingleScheduledTaskV1(t *testing.T) {
	const name = "hello"
	s := NewSchedulerV1()
	s.AddTask(name, 3*time.Second, func() {
		fmt.Println("run hello task")
	})

	time.Sleep(10 * time.Second)
	s.StopTask(name)
}

func TestRunMultiScheduledTaskV1(t *testing.T) {
	s := NewSchedulerV1()
	for i := 1; i <= 3; i++ {
		local := i
		name := fmt.Sprintf("task%d", local)
		s.AddTask(name, time.Duration(local)*time.Second, func() {
			fmt.Printf("run task by %d sec\n", local)
		})
	}

	time.Sleep(10 * time.Second)
	s.Close()
}

func TestRunSingleScheduledTaskV2(t *testing.T) {
	const name = "hello"
	s := NewSchedulerV2()
	s.AddTask(name, 3*time.Second, func() {
		fmt.Println("run hello task")
	})

	time.Sleep(10 * time.Second)
	s.StopTask(name)
}

func TestRunMultiScheduledTaskV2(t *testing.T) {
	s := NewSchedulerV2()
	for i := 1; i <= 3; i++ {
		local := i
		name := fmt.Sprintf("task%d", local)
		s.AddTask(name, time.Duration(local)*time.Second, func() {
			fmt.Printf("run task by %d sec\n", local)
		})
	}

	time.Sleep(10 * time.Second)
	s.Close()
}
