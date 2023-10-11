package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"demo.apps/utils"
	"github.com/gin-gonic/gin"
)

/*
page:
http://localhost:8081/static
http://localhost:8081/public/page_basic.html

rest api:
curl http://localhost:8081/
curl http://localhost:8081/ping

curl -v -XPOST http://localhost:8081/user -d '{"birthday":"10/07","timezone":"Asia/Shanghai"}'
*/

func main() {
	r := initServer()
	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}

// Router

func initServer() *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// refer: https://www.notefeel.com/you-trusted-all-proxies-this-is-not-safe
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})

	r.Use(gin.Recovery())

	r.NoMethod(HandleNotFound)
	r.NoRoute(HandleNotFound)

	r.GET("/", HandleIndex)
	r.GET("/ping", HandlePing)

	// validate middleware should be before CreateUser
	r.POST("/user", MiddlewareValidateJsonBody[CreateUserHttpBody](), HandleCreateUser)

	addStatic(r)
	addPagesStatic(r)

	return r
}

// Static

func addStatic(r *gin.Engine) {
	distPath := getDistPath()
	if utils.IsDirExist(distPath) {
		log.Println("static resource path:", distPath)
		r.Use(MiddlewareGzip())
		// Note: must add matched alias "/static/", "/static/index.html" for "/" in vue router.
		r.Static("/static", distPath)
		r.Static("/assets", filepath.Join(distPath, "assets"))
	}
}

func addPagesStatic(r *gin.Engine) {
	distPath := getPagesDistPath()
	if true { // for test
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		distPath = filepath.Join(dir, "public")
	}

	if utils.IsDirExist(distPath) {
		log.Println("static page path:", distPath)
		r.Use(func(c *gin.Context) {
			log.Println("hook: before send public page")
			c.Next()
			log.Println("hook: after send public page")
		})
		r.Static("/public", distPath)
	}
}

//nolint:unused
func addStaticEmbed(r *gin.Engine) {
	// TODO:
}

// Handle

func HandleNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func HandleIndex(c *gin.Context) {
	c.String(http.StatusOK, "gin server demo")
}

func HandlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// Demo: Create User Handle

type CreateUserHttpBody struct {
	Birthday string `json:"birthday" binding:"required,datetime=01/02"`
	Timezone string `json:"timezone" binding:"omitempty,timezone"`
}

func HandleCreateUser(c *gin.Context) {
	httpBody := GetJsonBody[CreateUserHttpBody](c)
	log.Printf("create user: birthday: %s, timezone: %s", httpBody.Birthday, httpBody.Timezone)
	c.JSON(http.StatusOK, gin.H{
		"message": "success created",
	})
}

const keyJsonBody = "jsonBody"

// MiddlewareValidateJsonBody validates request body by generic.
func MiddlewareValidateJsonBody[BodyType any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body BodyType
		if err := c.ShouldBindJSON(&body); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Set(keyJsonBody, body)
		c.Next()
	}
}

func GetJsonBody[BodyType any](c *gin.Context) BodyType {
	return c.MustGet(keyJsonBody).(BodyType)
}

// Middleware: gzip

// MiddlewareGzip gzip assets file if file size > 10kb.
func MiddlewareGzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if isAcceptEncodingGzip(c.Request.Header["Accept-Encoding"]) && isAssetsFilePath(path) {
			size, err := getFileSize(path)
			if err != nil {
				log.Println("get file size failed:", err)
				c.Next()
				return
			}

			if size > 10240 { // 10kb
				log.Println("add gzip encoding header for:", path)
				c.Header("Content-Encoding", "gzip")
			}
		}

		c.Next()
	}
}

// Helper

func isAcceptEncodingGzip(elems []string) bool {
	if len(elems) == 0 {
		return false
	}
	for _, elem := range elems {
		if strings.Contains(elem, "gzip") && strings.Contains(elem, "deflate") {
			return true
		}
	}
	return false
}

func isAssetsFilePath(url string) bool {
	return strings.HasPrefix(url, "/assets")
}

func getFileSize(relPath string) (int64, error) {
	fpath := filepath.Join(getDistPath(), relPath)
	stat, err := os.Stat(fpath)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

func getDistPath() string {
	const distRePath = "Workspaces/zj_repos/zj_js_project/vue3_lessons/demo_apps/app_basic/dist"
	return filepath.Join(os.Getenv("HOME"), distRePath)
}

func getPagesDistPath() string {
	const distRePath = "Workspaces/zj_repos/zj_js_project/vue3_lessons/demo_pages"
	return filepath.Join(os.Getenv("HOME"), distRePath)
}
