package demos

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestRandomInt(t *testing.T) {
	for i := 0; i < 30; i++ {
		rand.Seed(time.Now().UnixNano())
		fmt.Printf("%d,", rand.Int31n(100))
	}
	fmt.Println()
}

func TestTasksWeight(t *testing.T) {
	fn := func(idx int) {
		fmt.Printf("run task index=%d\t", idx)
	}

	// create weighted tasks
	var weightSum int32
	tasks := make([]*wTask, 0, 10)
	for _, weight := range []int32{10, 20, 30, 40} {
		tasks = append(tasks, &wTask{
			Weight: weight,
			Fn:     fn,
		})
		weightSum += weight
	}

	runTasksByWeight(tasks, 100*weightSum)

	fmt.Println("\nsummary:")
	for _, task := range tasks {
		fmt.Printf("task: weight=%d, count=%d\n", task.Weight, task.Count)
	}
}
