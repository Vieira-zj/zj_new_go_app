package demos

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

/*
Run tasks by diff weight
*/

type task struct {
	Count  int32 // task执行总次数
	Weight int32
	Fn     func(idx int)
}

func (t *task) Run(idx int) {
	atomic.AddInt32(&t.Count, 1)
	t.Fn(idx)
}

func TestRandomInt(t *testing.T) {
	for i := 0; i < 30; i++ {
		rand.Seed(time.Now().UnixNano())
		fmt.Printf("%d,", rand.Int31n(100))
	}
	fmt.Println()
}

func TestTasksWeight(t *testing.T) {
	tasks := make([]*task, 0, 10)

	fn := func(idx int) {
		fmt.Printf("run task index=%d\t", idx)
	}

	var weightSum int32
	for _, weight := range []int32{10, 20, 30, 40} {
		tasks = append(tasks, &task{
			Weight: weight,
			Fn:     fn,
		})
		weightSum += weight
	}

	ch := make(chan struct{}, 3)
	var wg sync.WaitGroup
	runCount := int(weightSum) * 100
	for i := 0; i < runCount; i++ {
		wg.Add(1)
		go func() {
			defer func() {
				<-ch
				wg.Done()
			}()

			rand.Seed(time.Now().UnixNano())
			val := rand.Int31n(weightSum)
			ch <- struct{}{}
			var base int32
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
