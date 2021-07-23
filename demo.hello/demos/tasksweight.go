package demos

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

/*
Run tasks by diff weights.
*/

type wTask struct {
	Count  int32 // task执行总次数
	Weight int32
	Fn     func(idx int)
}

func (t *wTask) Run(idx int) {
	atomic.AddInt32(&t.Count, 1)
	t.Fn(idx)
}

// runTasksByWeight runs tasks by totalRun times base on task weight.
func runTasksByWeight(tasks []*wTask, totalRun int32) {
	var weightSum int32
	for _, task := range tasks {
		weightSum += task.Weight
	}

	ch := make(chan struct{}, len(tasks))
	var wg sync.WaitGroup
	for i := 0; i < int(totalRun); i++ {
		wg.Add(1)
		go func() {
			defer func() {
				<-ch
				wg.Done()
			}()

			ch <- struct{}{}
			rand.Seed(time.Now().UnixNano())
			val := rand.Int31n(weightSum)
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
}
