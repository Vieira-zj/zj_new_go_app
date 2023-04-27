package main

import (
	"fmt"
	"math/rand"
	"time"
)

// 模拟 io 多路复用的场景
//
// - TimerFuture 代表一个 socket io 通道访问
// - Executor 执行器，监听和处理多个 io 通道
//

func init() {
	rand.Seed(time.Now().UnixMilli())
}

func main() {
	queue := make(chan Future)
	spawner := Spawner{
		queue: queue,
	}
	exec := Executor{
		queue: queue,
	}

	for i := 1; i <= 2; i++ {
		f := NewTimerFuture(i, queue)
		spawner.spawn(f)
		exec.incr()
	}

	exec.run()
	fmt.Println("run async done")
}
