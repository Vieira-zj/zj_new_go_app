package structs

import (
	"container/heap"
	"testing"
)

func TestPriorityQueue(t *testing.T) {
	tasks := []*Task{
		{ID: "task-low", Priority: 1},
		{ID: "task-high", Priority: 10},
		{ID: "task-medium", Priority: 5},
	}

	t.Run("priority queue with heap init", func(t *testing.T) {
		pq := make(PriorityQueue, 0, len(tasks))
		heap.Init(&pq)

		for _, task := range tasks {
			heap.Push(&pq, task)
		}
		for pq.Len() > 0 {
			task := heap.Pop(&pq).(*Task)
			t.Logf("processing task: %s (priority=%d)\n", task.ID, task.Priority)
		}
	})

	t.Run("priority queue without heap init", func(t *testing.T) {
		pq := make(PriorityQueue, 0, len(tasks))
		for _, task := range tasks {
			pq.Push(task)
		}
		for pq.Len() > 0 {
			task := pq.Pop().(*Task)
			t.Logf("processing task: %s (priority=%d)\n", task.ID, task.Priority)
		}
	})
}
