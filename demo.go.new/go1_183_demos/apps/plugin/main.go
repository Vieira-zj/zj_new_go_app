package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"plugin"
)

func main() {
	p, err := plugin.Open("./greeter.so")
	if err != nil {
		panic(err)
	}

	greeter, err := p.Lookup("Greeter")
	if err != nil {
		panic(err)
	}
	if res, ok := greeter.(*string); ok {
		fmt.Println("message:", *res)
	}

	greet, err := p.Lookup("Greet")
	if err != nil {
		panic(err)
	}
	if fn, ok := greet.(func(string) string); ok {
		fmt.Println(fn("Foo"))
	}

	ctx, stop := signal.NotifyContext(context.TODO(), os.Interrupt)
	<-ctx.Done()
	stop()

	fmt.Println("go plugin demo exit")
}
