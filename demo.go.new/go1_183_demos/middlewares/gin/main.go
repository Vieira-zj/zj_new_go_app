package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"demo.apps/utils"
	"github.com/gin-gonic/gin"
)

/*
page:
http://localhost:8081/static

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

	addStatic(r)

	r.GET("/", HandleIndex)
	r.GET("/ping", HandlePing)

	// validate middleware should be before CreateUser.
	r.POST("/user", ValidateJsonBody[CreateUserHttpBody](), HandleCreateUser)

	return r
}

// Static

func addStatic(r *gin.Engine) {
	distRePath := "Workspaces/zj_repos/zj_js_project/vue_apps/vue3_app_demo/dist"
	distPath := filepath.Join(os.Getenv("HOME"), distRePath)
	if utils.IsDirExist(distPath) {
		// Note: must add matched alias "/static/", "/static/index.html" for "/" in vue router.
		r.Static("/static", distPath)
		r.Static("/assets", filepath.Join(distPath, "assets"))
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

// Create User Handle

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

// ValidateJsonBody a middleware to validate request body by generic.
func ValidateJsonBody[BodyType any]() gin.HandlerFunc {
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
