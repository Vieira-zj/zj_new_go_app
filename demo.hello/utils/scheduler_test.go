package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestRunSingleScheduledTask(t *testing.T) {
	const name = "hello"
	s := NewScheduler()
	s.AddTask(3*time.Second, name, func() {
		fmt.Println("run hello task")
	})

	time.Sleep(10 * time.Second)
	s.StopTask(name)
}

func TestRunMultiScheduledTask(t *testing.T) {
	s := NewScheduler()
	for i := 1; i <= 3; i++ {
		local := i
		name := fmt.Sprintf("task%d", local)
		s.AddTask(time.Duration(local)*time.Second, name, func() {
			fmt.Printf("run task by %d sec\n", local)
		})
	}

	time.Sleep(10 * time.Second)
	s.Close()
}
