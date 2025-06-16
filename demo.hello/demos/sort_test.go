package demos

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGoroutinesPrint(t *testing.T) {
	// 使用N个goroutine交替打印出1-N, 比如N=3, 打印 123123123123...
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := make(chan int, 1)
	ch <- 1
	for i := 0; i < 3; i++ {
		go func() {
			for {
				select {
				case val := <-ch:
					fmt.Println(val)
					if val == 3 {
						val = 1
					} else {
						val++
					}
					time.Sleep(200 * time.Millisecond)
					ch <- val
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	<-ctx.Done()
	fmt.Println("done")
}
