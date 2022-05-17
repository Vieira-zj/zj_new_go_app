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
	rootDir string
	mode    string
	addr    string
	help    bool
)

func init() {
	flag.StringVar(&rootDir, "root", "/tmp/test/goc_plugin_space", "Goc plugin working root dir path.")
	flag.StringVar(&mode, "mode", "server", "Goc plugin run mode: server, watcher.")
	flag.StringVar(&addr, "addr", "8089", "Goc report server address.")
	flag.BoolVar(&help, "h", false, "help.")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	if err := pkg.InitConfig(rootDir); err != nil {
		log.Fatalf("Load config error: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	r := initRouter()
	switch mode {
	case modeServer:
		pkg.InitSrvCoverSyncTasksPool()
		defer pkg.CloseSrvCoverSyncTasksPool()
		runRptServer(ctx, r)
	case modeWatcher:
		runWatcher(ctx, r)
	default:
		log.Fatalln("Invalid run mode:", mode)
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
		c.String(http.StatusNotFound, fmt.Sprintf("Path not found: [%s] %s", c.Request.Method, c.Request.URL.Path))
	})

	r.GET("/", handler.IndexHandler)
	r.GET("/ping", handler.PingHandler)

	r.Static("/static/report", pkg.AppConfig.PublicDir)

	// middleware
	r.Use(gin.Logger())

	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", recovered))
	}))

	return r
}

//
// Report
//

func runRptServer(ctx context.Context, r *gin.Engine) {
	setupRptServerRouter(r)
	runRptScheduleTask(ctx)
}

func setupRptServerRouter(r *gin.Engine) {
	coverTotal := r.Group("/cover/total")
	coverTotal.GET("/list", handler.GetListOfSrvCoversHandler)
	coverTotal.POST("latest", handler.GetLatestSrvCoverTotalHandler)
	coverTotal.POST("history", handler.GetHistorySrvCoverTotalsHandler)

	cover := r.Group("/cover")
	cover.POST("raw", handler.FetchSrvRawCoverHandler)
	cover.POST("clear", handler.ClearSrvCoverHandler)
	cover.POST("sync", handler.SyncSrvCoverHandler)

	report := r.Group("/cover/report")
	report.POST("list", handler.ListSrvCoverReportsHandler)
	report.POST("download", handler.GetSrvCoverReportHandler)
}

func runRptScheduleTask(ctx context.Context) {
	scheduler := pkg.NewScheduler()
	scheduler.SyncRegisterSrvsCoverReportTask(ctx, 30*time.Minute)
}

//
// Watcher
//

func runWatcher(ctx context.Context, r *gin.Engine) {
	setupWatcherRouter(r)
	runWatcherScheduleTask(ctx)
}

func setupWatcherRouter(r *gin.Engine) {
	cover := r.Group("/watcher/cover")
	cover.POST("list", handler.ListSavedSrvCoversHandler)
	cover.POST("download", handler.GetSrvCoverHandler)
	cover.POST("sync", handler.FetchAndSaveSrvCoverHandler)
}

func runWatcherScheduleTask(ctx context.Context) {
	scheduler := pkg.NewScheduler()
	scheduler.RemoveUnhealthSrvTask(ctx, time.Hour)
	scheduler.SyncSrvsRawCoverTask(ctx, time.Hour)
}
