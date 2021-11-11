package main

import (
	"fmt"
	"testing"
	"time"
)

func TestBreakOuter(t *testing.T) {
loop:
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			fmt.Printf("%d-%d\n", i, j)
			if i == 2 && j == 3 {
				break loop
			}
		}
	}
	fmt.Println("done")
}

func TestCloseCh(t *testing.T) {
	go func() {
		for !cancelled() {
			fmt.Println("running...")
			time.Sleep(time.Second)
		}
		fmt.Println("exit")
	}()

	time.Sleep(time.Duration(3) * time.Second)
	close(cancelCh)
	time.Sleep(time.Duration(2) * time.Second)
	fmt.Println("done")
}
