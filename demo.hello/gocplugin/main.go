package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"demo.hello/gocplugin/pkg"
	"demo.hello/gocplugin/pkg/handler"
	"github.com/gin-gonic/gin"
)

const (
	modeServer  = "server"
	modeWatcher = "watcher"
)

var (
	cfgPath string
	mode    string
	addr    string
	help    bool
)

func init() {
	flag.StringVar(&cfgPath, "c", "/tmp/test/gocplugin.json", "Goc plugin config file path.")
	flag.StringVar(&mode, "mode", "server", "Goc plugin run mode: server, watcher.")
	flag.StringVar(&addr, "addr", "8089", "Goc server address.")
	flag.BoolVar(&help, "h", false, "help.")
	flag.Parse()
}

func main() {
	if help {
		flag.Usage()
		return
	}

	if err := pkg.LoadConfig(cfgPath); err != nil {
		log.Fatalf("Load config error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	pkg.InitSrvCoverSyncTasksPool()
	defer pkg.CloseSrvCoverSyncTasksPool()

	r := initRouter()
	switch mode {
	case modeServer:
		setupServerRouter(r)
	case modeWatcher:
		setupWatcherRouter(r)
		runWatcherScheduleTask(ctx)
	default:
		log.Fatalln("Invalid mode:", mode)
	}

	go func() {
		if err := r.Run(":" + addr); err != nil {
			log.Fatalln(err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Println("Goc plugin exit.")
}

func initRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, fmt.Sprintf("Path not found: %s", c.Request.URL.Path))
	})

	r.GET("/", handler.IndexHandler)
	r.GET("/ping", handler.PingHandler)
	return r
}

func setupServerRouter(r *gin.Engine) {
	r.GET("/cover/list", handler.GetListOfSrvCoversHandler)
	r.POST("/cover/total/latest", handler.GetLatestSrvCoverTotalHandler)
	r.POST("/cover/total/history", handler.GetHistorySrvCoverTotalsHandler)

	r.POST("/cover/report/sync", handler.SyncSrvCoverHandler)

	r.POST("/cover/raw", handler.GetSrvRawCoverHandler)
	r.POST("/cover/report/latest", handler.GetLatestSrvCoverReportHandler)

	// middleware
	r.Use(gin.Logger())

	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", recovered))
	}))
}

func setupWatcherRouter(r *gin.Engine) {
	// TODO:
}

func runWatcherScheduleTask(ctx context.Context) {
	scheduler := pkg.NewScheduler()
	scheduler.RemoveUnhealthSrvTask(ctx, 10*time.Second)
	scheduler.SyncRegisterSrvsCoverTask(ctx, 10*time.Second)
}
