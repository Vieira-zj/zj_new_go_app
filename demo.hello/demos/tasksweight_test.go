package demos

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
)

/*
Run tasks by diff weight
*/

type task struct {
	Count  int32
	Weight int32
	Fn     func(idx int)
}

func (t *task) Run(idx int) {
	atomic.AddInt32(&t.Count, 1)
	t.Fn(idx)
}

func TestTasksWeight(t *testing.T) {
	tasks := make([]*task, 0, 10)

	fn := func(idx int) {
		fmt.Printf("run task index=%d\t", idx)
	}

	var weightSum int32
	for _, weight := range []int32{10, 20, 30} {
		tasks = append(tasks, &task{
			Weight: weight,
			Fn:     fn,
		})
		weightSum += weight
	}

	ch := make(chan struct{}, 3) // parallel number
	var wg sync.WaitGroup
	runCount := int(weightSum) * 100
	for i := 0; i < runCount; i++ {
		wg.Add(1)
		go func() {
			defer func() {
				<-ch
				wg.Done()
			}()

			ch <- struct{}{}
			var base int32
			val := rand.Int31n(weightSum)
			for idx, task := range tasks {
				base += task.Weight
				if val < base {
					task.Run(idx)
					break
				}
			}
		}()
	}
	wg.Wait()

	fmt.Println("\nsummary:")
	for _, task := range tasks {
		fmt.Printf("task: weight=%d, count=%d\n", task.Weight, task.Count)
	}
}
