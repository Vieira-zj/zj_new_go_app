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
		runServer(r)
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

func runServer(r *gin.Engine) {
	setupServerRouter(r)
}

func setupServerRouter(r *gin.Engine) {
	r.GET("/goc/list/sync", handler.SyncGocListHandler)

	coverTotal := r.Group("/cover/total")
	coverTotal.GET("/list", handler.GetListOfSrvCoversHandler)
	coverTotal.POST("latest", handler.GetLatestSrvCoverTotalHandler)
	coverTotal.POST("history", handler.GetHistorySrvCoverTotalsHandler)

	cover := r.Group("/cover")
	cover.POST("sync", handler.SyncSrvCoverHandler)
	cover.POST("clear", handler.ClearSrvCoverHandler)

	report := r.Group("/cover/report")
	report.POST("list", handler.ListSrvCoverReportsHandler)
	report.POST("raw", handler.GetSrvRawCoverHandler)
	report.POST("func", handler.GetSrvFuncCoverReportHandler)
}

func runWatcher(ctx context.Context, r *gin.Engine) {
	setupWatcherRouter(r)
	runWatcherScheduleTask(ctx)
}

func setupWatcherRouter(r *gin.Engine) {
	cover := r.Group("/watcher/cover")
	cover.POST("list", handler.ListSavedSrvCoversHandler)
	cover.POST("get", handler.GetSrvCoverDataHandler)
	cover.POST("save", handler.FetchAndSaveSrvCoverHandler)
}

func runWatcherScheduleTask(ctx context.Context) {
	scheduler := pkg.NewScheduler()
	scheduler.RemoveUnhealthSrvTask(ctx, time.Hour)
	scheduler.SyncRegisterSrvsCoverTask(ctx, time.Hour)
}
