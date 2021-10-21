package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	k8spkg "demo.hello/k8s/client/pkg"
	"demo.hello/k8s/monitor/handlers"
	"demo.hello/k8s/monitor/internal"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

var (
	addr      string
	namespace string
	interval  uint
	help      bool
)

func main() {
	flag.StringVar(&addr, "addr", "8081", "server listen port.")
	flag.StringVar(&namespace, "ns", "k8s-test", "monitor namespace.")
	flag.UintVar(&interval, "interval", 15, "interval (seconds) between send notify message.")
	flag.BoolVar(&help, "h", false, "help.")
	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	// run list watcher
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	client, err := k8spkg.CreateK8sClientLocalDefault()
	if err != nil {
		panic(err)
	}
	watcher := internal.NewWatcher(client, namespace, interval)
	lister := internal.NewLister(client, watcher, namespace)
	watcher.Run(ctx, false)

	// run server
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	e.GET("/", deco(handlers.Home))
	e.GET("/ping", deco(handlers.Ping))

	e.GET("/resource/pods", deco(func(c echo.Context) error {
		return handlers.GetPodsStatus(c, lister)
	}))
	e.GET("/list/pods", deco(func(c echo.Context) error {
		return handlers.GetPodsStatusByList(c, lister)
	}))

	go func() {
		addr = fmt.Sprintf(":%s", addr)
		e.Logger.Info("Start cluster monitor at " + addr)
		e.Logger.Fatal(e.Start(addr))
	}()

	// send notify
	go func() {
		tick := time.Tick(time.Duration(interval) * time.Second)
		for {
			select {
			case <-tick:
				for {
					if len(watcher.ErrorPodStateCh) == 0 {
						break
					}
					state := <-watcher.ErrorPodStateCh
					b, err := json.MarshalIndent(state, "", "  ")
					if err != nil {
						e.Logger.Errorf("json marshal error: %v\n", err)
					}
					e.Logger.Info("[Notification] pod not running:")
					e.Logger.Info(string(b))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	stop()

	// clearup
	if err := e.Close(); err != nil {
		panic(err)
	}
	e.Logger.Info("k8s monitor done")
}

func deco(fn func(echo.Context) error) func(echo.Context) error {
	return func(c echo.Context) error {
		preHook(c)
		err := fn(c)
		afterHook(c)
		return err
	}
}

func preHook(c echo.Context) {
	request := c.Request()
	c.Logger().Info("| Host: ", request.Host)
	c.Logger().Info("| Url: ", request.URL)
	c.Logger().Info("| Method: ", request.Method)
}

func afterHook(c echo.Context) {
}
