package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	logs "log"
	"os"
	"os/signal"
	"strings"
	"time"

	k8spkg "demo.hello/k8s/client/pkg"
	"demo.hello/k8s/monitor/handlers"
	"demo.hello/k8s/monitor/internal"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

var (
	addr     string
	ns       string
	interval uint
	help     bool
)

func main() {
	flag.StringVar(&addr, "addr", "8081", "http server listen port.")
	flag.StringVar(&ns, "ns", "k8s-test,default", "target namespaces to be monitor, split by ','.")
	flag.UintVar(&interval, "interval", 15, "interval (seconds) for list watcher to sync data.")
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
	namespaces := strings.Split(ns, ",")
	watcher := internal.NewWatcher(client, namespaces, interval)
	lister := internal.NewLister(client, watcher, namespaces)
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
		logs.Println("Start cluster monitor at " + addr)
		logs.Fatalln(e.Start(addr))
	}()

	// send notify
	mm, err := internal.NewMatterMost()
	if err != nil {
		panic(err)
	}
	go func() {
		tick := time.Tick(time.Duration(interval*3) * time.Second)
		for {
			select {
			case <-tick:
				statusMap := make(map[string]*internal.PodStatus, interval)
				for {
					if len(watcher.ErrorPodStatusCh) == 0 {
						break
					}
					status := <-watcher.ErrorPodStatusCh
					statusMap[status.Name] = status // distinct
				}
				for _, status := range statusMap {
					notify(ctx, mm, status)
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
	logs.Println("k8s monitor done")
}

func notify(ctx context.Context, mm *internal.MatterMost, status *internal.PodStatus) {
	podLog := status.Log
	status.Log = ""
	b, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		logs.Printf("json marshal error: %v\n", err)
	}

	logs.Println("send notification")
	defaultUser := "jin.zheng"
	msg := fmt.Sprintf("`Notification` pods not running:\n%s", markdownBlockText("json", string(b)))
	mm.SendMessageToUser(ctx, defaultUser, msg)
	if len(podLog) > 0 {
		msg = fmt.Sprintf("Pod `%s` Log:\n%s", status.Name, markdownBlockText("text", podLog))
		mm.SendMessageToUser(ctx, defaultUser, msg)
	}
}

func markdownBlockText(mdType, text string) string {
	return fmt.Sprintf("```%s\n%s\n```", mdType, text)
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
