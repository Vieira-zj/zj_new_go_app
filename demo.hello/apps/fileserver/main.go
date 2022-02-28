package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"demo.hello/utils"
	"github.com/gin-gonic/gin"
)

const (
	publicLintDir = "./public/lint"
)

var (
	port    string
	expired float64
	modules string
	help    bool
)

func init() {
	flag.StringVar(&port, "p", ":8081", "Server listen port.")
	flag.Float64Var(&expired, "e", 24, "Clear history files whose mod time gt expired time.")
	flag.StringVar(&modules, "m", "spba", "Modules to be register in static router.")
	flag.BoolVar(&help, "h", false, "Help.")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	r := setupRouter()
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Println("Server closed")
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	<-ctx.Done()
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server forced to shutdown:", err)
	}
}

//
// Gin Router
//

func setupRouter() *gin.Engine {
	r := gin.Default()
	registerStaticRouter(r)

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello Gin")
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/upload", fileUploadHandler)

	// register static router
	r.GET("/register", func(c *gin.Context) {
		module := c.Query("module")
		if len(module) == 0 {
			c.String(http.StatusBadRequest, "Param [module] is not set")
			return
		}
		registerStaticDir(r, module)
		c.String(http.StatusOK, fmt.Sprintf("Module [%s] registered", module))
	})

	return r
}

func registerStaticRouter(r *gin.Engine) {
	mdls := strings.Split(modules, ",")
	for _, module := range mdls {
		registerStaticDir(r, module)
	}
}

func registerStaticDir(r *gin.Engine, module string) {
	path := fmt.Sprintf("%s/%s", publicLintDir, module)
	r.Static(path, path)
}

func fileUploadHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusInternalServerError, "Read file error:", err)
		return
	}

	component := strings.ToLower(c.GetHeader("X-Component"))
	fileSavePath, err := getFileSavePath(component, file.Filename)
	if err != nil {
		c.String(http.StatusInternalServerError, "Get file save path error:", err)
		return
	}

	if err := c.SaveUploadedFile(file, fileSavePath); err != nil {
		c.String(http.StatusInternalServerError, "Save file error:", err)
		return
	}
	c.String(http.StatusOK, fmt.Sprintf("[%s] uploaded", file.Filename))
}

func getFileSavePath(component, fileName string) (string, error) {
	saveDir := filepath.Join(publicLintDir, component)
	if !utils.IsDirExist(saveDir) {
		if err := utils.MakeDir(saveDir); err != nil {
			return "", nil
		}
	} else {
		// 删除过期文件
		go func() {
			if err := clearExpiredFiles(saveDir); err != nil {
				log.Println("Clear expired files error:", err)
			}
		}()
	}
	return filepath.Join(saveDir, fileName), nil
}

func clearExpiredFiles(dirPath string) error {
	files, err := utils.RemoveExpiredFiles(dirPath, expired, utils.Hour)
	if err != nil {
		return err
	}
	if len(files) > 0 {
		log.Printf("Clear expired files from dir [%s]: %s\n", dirPath, strings.Join(files, ","))
	}
	return nil
}

//
// Gin Middleware
//

func setupMiddleware(r *gin.Engine) {
	r.Use(gin.Logger())

	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintln("Internal error:", err))
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
}
