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
	"syscall"
	"time"

	k8spkg "demo.hello/k8s/client/pkg"
	"demo.hello/k8s/monitor/handlers"
	"demo.hello/k8s/monitor/internal"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"k8s.io/client-go/kubernetes"
)

var (
	addr, runMode, ns string
	isDebug, help     bool
	interval          uint
	duration          int
)

func init() {
	// interval time:
	// 1. watcher sync data by "interval"
	// 2. message queue size is set to "interval"
	// 3. get message from queue and send notification with "3*interval"
	// 4. send notification with rate limiter [burst=3, rate=1] per duration minutes.

	flag.StringVar(&addr, "addr", "8081", "http server listen port.")
	flag.BoolVar(&isDebug, "debug", false, "debug mode, default false.")
	flag.StringVar(&runMode, "mode", "local", "k8s monitor run mode: local, cluster. defalut: local.")
	flag.StringVar(&ns, "ns", "default", "target list of namespaces to be monitor, split by ','.")
	flag.UintVar(&interval, "interval", 15, "interval (seconds) for list watcher to sync data.")
	flag.IntVar(&duration, "duration", 5, "rate limiter duration (minutes) for send notification.")
	flag.BoolVar(&help, "h", false, "help.")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	// init list watcher
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)

	client, err := initK8sClient()
	if err != nil {
		panic(fmt.Sprintf("init k8s client error: %v", err))
	}
	namespaces := strings.Split(ns, ",")
	watcher := internal.NewWatcher(client, namespaces, interval, isDebug)
	lister := internal.NewLister(client, watcher, namespaces)
	if err := watcher.Run(ctx, false); err != nil {
		panic(fmt.Sprintf("run watcher error: %v", err))
	}

	// run http server
	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	initServerRouter(e, lister)

	go func() {
		addr = fmt.Sprintf(":%s", addr)
		logs.Println("Start cluster monitor at " + addr)
		logs.Fatalln(e.Start(addr))
	}()

	// run ratelimiter
	limiter := internal.NewRateLimiter(duration*60, 3)
	go func() {
		limiter.Run(ctx)
	}()

	// run notify
	mm := internal.NewMatterMost()
	go func() {
		tick := time.Tick(time.Duration(3*interval) * time.Second)
		for {
			select {
			case <-tick:
				handler(ctx, watcher, limiter, mm)
			case <-ctx.Done():
				logs.Println("notify exit")
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
	logs.Println("k8s monitor exit")
}

//
// common
//

func initK8sClient() (*kubernetes.Clientset, error) {
	if strings.ToLower(runMode) == "local" {
		return k8spkg.CreateK8sClientLocalDefault()
	}
	return k8spkg.CreateK8sClient()
}

func initServerRouter(e *echo.Echo, lister *internal.Lister) {
	e.GET("/", deco(handlers.Index))
	e.GET("/ping", deco(handlers.Ping))

	e.GET("/resource/pods", deco(func(c echo.Context) error {
		return handlers.GetPodsStatus(c, lister)
	}))
	e.GET("/list/pods", deco(func(c echo.Context) error {
		return handlers.GetPodsStatusByList(c, lister)
	}))
	e.POST("/list/pods/filter", deco(func(c echo.Context) error {
		return handlers.GetPodsStatusByFilter(c, lister)
	}))
}

func handler(ctx context.Context, watcher *internal.Watcher, limiter *internal.RateLimiter, mm *internal.MatterMost) {
	statusMap := make(map[string]*internal.PodStatus, interval) // distinct values
outer:
	for {
		select {
		case status := <-watcher.ErrorPodStatusCh:
			statusMap[status.Name] = status
		default:
			break outer
		}
	}
	for _, status := range statusMap {
		if !limiter.Add(status.Name) {
			logs.Printf("exceed the rate limit, ignore: [namespace=%s,name=%s,status=%s]",
				status.Namespace, status.Name, status.Value)
			continue
		}
		notify(ctx, mm, status)
	}
}

func notify(ctx context.Context, mm *internal.MatterMost, status *internal.PodStatus) {
	// 日志中包含换行，单独输出
	podLog := status.Log
	status.Log = ""
	b, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		logs.Printf("json marshal error: %v\n", err)
	}

	logs.Println("send notification")
	defaultUser := "jin.zheng"
	msg := fmt.Sprintf("`Notification:` pods not running:\n%s", markdownBlockText("json", string(b)))
	mm.SendMessageToUser(ctx, defaultUser, msg)
	if len(podLog) > 0 {
		msg = fmt.Sprintf("Pod `%s` Log:\n%s", status.Name, markdownBlockText("text", podLog))
		mm.SendMessageToUser(ctx, "", msg)
	}
}

func markdownBlockText(textType, text string) string {
	return fmt.Sprintf("```%s\n%s\n```", textType, text)
}

//
// http hooks
//

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
