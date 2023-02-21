package main

import (
	"bytes"
	"fmt"
	"go1_1711_demo/middlewares/gin/pkg"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func notFoundHandler(c *gin.Context) {
	// c.FullPath() returns empty string here
	log.Printf("full_path:%s, url_path:%s", c.FullPath(), c.Request.URL.Path)
	err := pkg.NotFound
	c.JSON(err.StatusCode, err)
	return
}

func pingHandler(c *gin.Context) {
	time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{
		"message": "Pong",
	})
}

func loginHandler(c *gin.Context) {
	log.Println("access login")
	name := c.Query("name")
	c.String(http.StatusOK, "welcome "+name)
}

func panicHandler(c *gin.Context) {
	trigger := false
	if tri, ok := c.GetQuery("trigger"); ok && tri == "true" {
		trigger = true
	}

	if trigger {
		log.Println("trigger panic")
		panic("mock panic")
	}

	c.String(http.StatusOK, "no trigger panic")
}

// CopyBody Handler

type testBody struct {
	Id      int    `json:"id" binding:"required"`
	Content string `json:"content"`
}

func copyBodyHandler(c *gin.Context) {
	body := c.Request.Body
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "read request body error: " + err.Error(),
		})
		return
	}
	log.Printf("post body: %s", bodyBytes)
	body.Close()

	// reset body
	c.Request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	req := testBody{}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parse json body error: " + err.Error(),
		})
		return
	}
	log.Printf("read json body: id=%d, content=[%s]", req.Id, req.Content)

	c.JSON(http.StatusOK, gin.H{
		"message": "cpoied ok",
	})
}

// Chunked Handler

func streamHandler(c *gin.Context) {
	w := c.Writer
	headers := w.Header()
	headers.Set("Transfer-Encoding", "chunked")
	headers.Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	// header
	w.Write([]byte("<html>\n  <body>\n"))
	w.Flush()

	// body
	for i := 0; i < 6; i++ {
		w.Write([]byte(fmt.Sprintf("    <h1>%d</h1>\n", i)))
		w.Flush()
		time.Sleep(time.Second)
	}

	// tailer
	w.Write([]byte("  </body>\n</html>"))
	w.Flush()
}

//
// Abort Handler
//
// normal:       logger1 -> abort -> logger2 -> handler -> logger2 -> abort -> logger1
// before abort: logger1 -> abort -> loggeer1
// after abort:  logger1 -> abort -> logger2 -> handler -> logger2 -> logger1
//

func abortHandler(c *gin.Context) {
	log.Println("abort handler go")

	time.Sleep(100 * time.Millisecond)
	abortType := c.Request.URL.Query().Get("type")
	if abortType == "after" {
		c.Error(fmt.Errorf("mock error: abort after"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "abort handler finished",
	})
}

// Mock Error Handler

func userHandler(c *gin.Context) {
	name := c.Param("name")
	c.Error(fmt.Errorf(fmt.Sprintf("user name [%s] not found", name)))
}

func usersHandler(c *gin.Context) {
	c.Error(pkg.ServerError)
}

// Param Validate Handler

type signUpParam struct {
	Age   uint8  `json:"age" binding:"gte=1,lte=130"`
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Date  string `json:"date" binding:"required,datetime=2006-01-02,checkDate"`
}

func checkSignUpDate(fl validator.FieldLevel) bool {
	date, err := time.Parse("2006-01-02", fl.Field().String())
	if err != nil {
		return false
	}
	if date.Before(time.Now()) {
		return false
	}
	return true
}

func signUpHanler(c *gin.Context) {
	var s signUpParam
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Printf("sign up: name=%s, age=%d, email=%s, date=%s", s.Name, s.Age, s.Email, s.Date)
	c.JSON(http.StatusOK, gin.H{
		"message": "signup success",
	})
}

// Prometheus Handler

func prometheusHandlerA(c *gin.Context) {
	sleep := rand.Intn(200)
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "apia get ok",
	})
}

func prometheusHandlerB(c *gin.Context) {
	sleep := rand.Intn(300)
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "apib post ok",
	})
}

// Wrap Context Handler

const ctxTestKey = "ctx-test-key"

func getContextValueHandler(c *gin.Context) {
	// condition: enable r.ContextWithFallback
	val := c.Value(ctxTestKey)
	if val == nil {
		log.Println("no value in gin context")
	}
	log.Println("value in gin context:", val)

	ctx := c.Request.Context()
	val = ctx.Value(ctxTestKey)
	if val == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no value in request context",
		})
		return
	}
	log.Println("value in request context:", val)

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"ctx-value": val,
	})
}
