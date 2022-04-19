package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"demo.hello/gocplugin/pkg"
	"demo.hello/gocplugin/pkg/handler"
	"github.com/gin-gonic/gin"
)

var (
	cfgPath string
	addr    string
	help    bool
)

func init() {
	flag.StringVar(&cfgPath, "c", "/tmp/test/gocplugin.json", "Goc watch dog config file path.")
	flag.StringVar(&addr, "addr", ":8089", "Goc server address.")
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

	pkg.InitSrvCoverSyncTasksPool()
	defer pkg.CloseSrvCoverSyncTasksPool()

	r := setupRouter()
	if err := r.Run(addr); err != nil {
		log.Fatalln(err)
	}
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// route
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, fmt.Sprintf("Path not found: %s", c.Request.URL.Path))
	})

	r.GET("/", handler.IndexHandler)
	r.GET("/ping", handler.PingHandler)
	r.POST("/cover/raw", handler.GetSrvRawCoverHandler)

	r.GET("/cover/list", handler.GetCoverSrvListHandler)
	r.POST("/cover/report/sync", handler.SyncSrvCoverHandler)
	r.POST("cover/latest/report", handler.GetLatestCoverReportHandler)

	// middleware
	r.Use(gin.Logger())

	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", recovered))
	}))
	return r
}
