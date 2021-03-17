package main

import (
	"fmt"
	"time"

	"demo.hello/apps/reversecall/pkg/test/example"
)

func main() {
	fmt.Println("start")
	example.Test3()
	example.Test3a()
	example.Test3c()
	go example.ReceiveFromKafka()
	time.Sleep(time.Second)
}
