package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"demo.apps/utils"
	"github.com/gin-gonic/gin"
)

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
	r.Any("/echo", HandleEcho)

	r.POST("/upload", HandleUpload)

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

// Default Handle

func HandleNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, http.StatusText(http.StatusNotFound))
}

func HandleIndex(c *gin.Context) {
	// test get env var from debug session
	if val, ok := os.LookupEnv("X-DEBUG"); ok {
		log.Println("debug mode:", val)
	}
	c.String(http.StatusOK, "gin server demo")
}

func HandlePing(c *gin.Context) {
	// only write header
	c.Writer.WriteHeader(http.StatusOK)
}

func HandleEcho(c *gin.Context) {
	log.Println("method:", c.Request.Method)
	log.Println("host:", c.Request.Host)

	log.Println("query:")
	for k, v := range c.Request.URL.Query() {
		fmt.Printf("\tkey=%s, value=%s\n", k, v)
	}

	log.Println("headers:")
	for k, v := range c.Request.Header {
		fmt.Printf("\tkey=%s, value=%s\n", k, v)
	}

	body, err := io.ReadAll(c.Request.Body)
	defer c.Request.Body.Close()
	if err != nil {
		log.Println("read body error:", err)
	}
	if len(body) > 0 {
		log.Println("body:")
		fmt.Println(string(body))
	}

	if timeout := c.Query("timeout"); len(timeout) > 0 {
		duration, err := time.ParseDuration(timeout)
		if err != nil {
			log.Println("invalid timeout:", timeout)
			c.Writer.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Println("sleep for " + timeout)
		time.Sleep(duration)
	}

	c.Writer.WriteHeader(http.StatusOK)
}

// Handler: Upload File

func HandleUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Note: key "files" matches the name in "curl F/--form <name=content>".
	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "no multipart file header found",
		})
		return
	}

	for _, f := range files {
		fpath := fmt.Sprintf("/tmp/test/%d_%s", time.Now().UnixMilli(), f.Filename)
		log.Println("save upload file to: " + fpath)
		if err = c.SaveUploadedFile(f, fpath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "save upload file error: " + err.Error(),
			})
			return
		}

		md5hash := c.GetHeader("X-Md5")
		if len(md5hash) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "no file md5 hash provided",
			})
			return
		}

		ok, err := verifyFileMd5hash(fpath, md5hash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		if !ok {
			log.Printf("verify upload file [%s] md5 hash failed", f.Filename)
			c.JSON(http.StatusOK, gin.H{
				"error": "save upload file error: md5 hash is not matched",
			})
			return
		}
	}

	c.Writer.WriteHeader(http.StatusOK)
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
