package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	addr := flag.String("h", "127.0.0.1:6379", "redis host")
	uname := flag.String("u", "", "username")
	pwd := flag.String("p", "", "password")
	flag.Parse()

	client := getRedisClient(*addr, *uname, *pwd)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	go func() {
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println(ctx.Err())
				return
			case <-t.C:
			}

			result, err := client.Ping(context.TODO()).Result()
			if err != nil {
				fmt.Println("ping error:", err)
			} else {
				fmt.Println("ping output:", result)
			}
		}
	}()

	<-ctx.Done()
	cancel()

	time.Sleep(300 * time.Millisecond)
	fmt.Println("exit")
}

func getRedisClient(addr, uname, pwd string) *redis.Client {
	opt := redis.Options{
		Addr: addr,
	}
	if len(uname) > 0 {
		opt.Username = uname
	}
	if len(pwd) > 0 {
		opt.Password = pwd
	}

	return redis.NewClient(&opt)
}
