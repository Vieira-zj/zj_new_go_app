package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	logs "log"
	"net/http"
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

const (
	defaultUser = "jin.zheng"
)

var (
	addr, runMode, ns string
	help, isDebug     bool
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)

	// init list watcher
	client := initK8sClient()
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

	stateHandler := func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"is_notify":  handlers.IsNotify,
			"queue_size": len(watcher.ErrorPodStatusCh),
		})
	}
	e.GET("/state", deco(stateHandler))

	go func() {
		addr = fmt.Sprintf(":%s", addr)
		logs.Println("Start cluster monitor at " + addr)
		logs.Fatalln(e.Start(addr))
	}()

	// run ratelimiter
	limiter := internal.NewRateLimiter(duration*60, 3)
	limiter.Start()
	defer limiter.Stop()

	// run notify
	go func() {
		tick := time.Tick(time.Duration(3*interval) * time.Second)
		for {
			select {
			case <-tick:
				notifyProcess(ctx, watcher, limiter)
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

func initK8sClient() *kubernetes.Clientset {
	var (
		client *kubernetes.Clientset
		err    error
	)
	if strings.ToLower(runMode) == "local" {
		if client, err = k8spkg.CreateK8sClientLocalDefault(); err != nil {
			msg := "init k8s local error: " + err.Error()
			panic(msg)
		}
	} else {
		if client, err = k8spkg.CreateK8sClient(); err != nil {
			msg := "init k8s cluster error: " + err.Error()
			panic(msg)
		}
	}
	return client
}

func initServerRouter(e *echo.Echo, lister *internal.Lister) {
	e.GET("/", deco(handlers.Index))
	e.GET("/ping", deco(handlers.Ping))
	e.GET("/notify", deco(handlers.SetNotify))

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

func notifyProcess(ctx context.Context, watcher *internal.Watcher, limiter *internal.RateLimiter) {
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
		if limiter.Acquire(status.Name) {
			notifyToChannel(ctx, status)
		} else {
			logs.Printf("exceed the rate limit, ignore: [namespace=%s,name=%s,status=%s]",
				status.Namespace, status.Name, status.Value)
		}
	}
}

func notifyToChannel(ctx context.Context, status *internal.PodStatus) {
	// 日志中包含换行，单独输出
	podLog := status.Log
	status.Log = ""
	b, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		logs.Printf("json marshal error: %v", err)
	}

	logs.Println("send notification")
	msg := fmt.Sprintf("`Notification:` pods not running:\n%s", markdownBlockText("json", string(b)))
	sendNotifyAtUser(ctx, msg)
	if len(podLog) > 0 {
		msg = fmt.Sprintf("Pod `%s` Log:\n%s", status.Name, markdownBlockText("text", podLog))
		sendNotify(ctx, msg)
	}
}

func sendNotifyAtUser(ctx context.Context, msg string) {
	if !handlers.IsNotify {
		logs.Printf("log notify:\n%s", msg)
		return
	}
	mm := internal.NewMatterMost()
	if err := mm.SendMessageToUser(ctx, defaultUser, msg); err != nil {
		logs.Printf("send notify error: %v", err)
	}
}

func sendNotify(ctx context.Context, msg string) {
	if !handlers.IsNotify {
		logs.Printf("log notify:\n%s", msg)
		return
	}
	mm := internal.NewMatterMost()
	if err := mm.SendMessageToUser(ctx, "", msg); err != nil {
		logs.Printf("send notify error: %v", err)
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
