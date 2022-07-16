package main

import (
	"fmt"
	"time"

	sdkRedis "go1_1711_demo/ioc.demo/autowire_redis_client/sdk"

	"github.com/alibaba/ioc-golang"
	"github.com/alibaba/ioc-golang/config"
	"github.com/go-redis/redis"
)

// +ioc:autowire=true
// +ioc:autowire:type=singleton
// +ioc:autowire:paramType=Param
// +ioc:autowire:constructFunc=Init
// +ioc:autowire:alias=AppAlias

type App struct {
	NormalRedis    sdkRedis.RedisIOCInterface `normal:""`
	NormalDB1Redis sdkRedis.RedisIOCInterface `normal:",db1-redis"`
	NormalDB2Redis sdkRedis.RedisIOCInterface `normal:",db2-redis"`
	NormalDB3Redis sdkRedis.RedisIOCInterface `normal:",address=127.0.0.1:6379&db=3"`

	nonIocClient *redis.Client
}

type Param struct {
	RedisAddr string
	DB        int
}

func (p *Param) Init(a *App) (*App, error) {
	client := redis.NewClient(&redis.Options{
		Addr: p.RedisAddr,
		DB:   p.DB,
	})
	a.nonIocClient = client
	return a, nil
}

func (a *App) Run() {
	if _, err := a.NormalRedis.Set("mykey", "db0", -1).Result(); err != nil {
		panic(err)
	}
	if _, err := a.NormalDB1Redis.Set("mykey", "db1", -1).Result(); err != nil {
		panic(err)
	}
	if _, err := a.NormalDB2Redis.Set("mykey", "db2", -1).Result(); err != nil {
		panic(err)
	}
	if _, err := a.NormalDB3Redis.Set("mykey", "db3", -1).Result(); err != nil {
		panic(err)
	}
	if status := a.nonIocClient.Set("mykey", "db4", -1); status.Err() != nil {
		panic(status.Err())
	}

	for {
		time.Sleep(3 * time.Second)
		val1, err := a.NormalRedis.Get("mykey").Result()
		if err != nil {
			panic(err)
		}
		fmt.Println("client0 get ", val1)

		val2, err := a.NormalDB1Redis.Get("mykey").Result()
		if err != nil {
			panic(err)
		}
		fmt.Println("client1 get ", val2)

		val3, err := a.NormalDB2Redis.Get("mykey").Result()
		if err != nil {
			panic(err)
		}
		fmt.Println("client2 get ", val3)

		val4, err := a.NormalDB3Redis.Get("mykey").Result()
		if err != nil {
			panic(err)
		}
		fmt.Println("client3 get ", val4)

		status := a.nonIocClient.Get("mykey")
		if status.Err() != nil {
			panic(err)
		}
		fmt.Println("non-ioc client get:", status.Val())
	}
}

func main() {
	if err := ioc.Load(
		config.WithSearchPath("../conf"),
		config.WithConfigName("ioc_golang")); err != nil {
		panic(err)
	}

	app, err := GetAppSingleton(&Param{
		RedisAddr: "localhost:6379",
		DB:        4,
	})
	if err != nil {
		panic(err)
	}
	app.Run()
}
