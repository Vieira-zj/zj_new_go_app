package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleIndex(t *testing.T) {
	w := httptest.NewRecorder()
	c := GetTestGinContext(w)

	HandleIndex(c)
	t.Log("ret code:", w.Code)
	t.Log("resp body:", w.Body.String())
}

func TestHandleCreateUser(t *testing.T) {
	w := httptest.NewRecorder()
	c := GetTestGinContext(w)

	b := []byte(`{"birthday":"10/07","timezone":"Asia/Shanghai"}`)
	SetTestWriteBodyforPost(c, b)

	// as set context in middleware
	var body CreateUserHttpBody
	if err := c.ShouldBindJSON(&body); err != nil {
		t.Fatal(err)
	}
	c.Set(keyJsonBody, body)

	HandleCreateUser(c)
	t.Log("ret code:", w.Code)
	t.Log("resp body:", w.Body.String())
}

// Gin Test Helper

func GetTestGinContext(w *httptest.ResponseRecorder) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Header: make(http.Header),
		URL:    &url.URL{},
	}
	return c
}

func SetTestUrlQueryForGet(c *gin.Context, pairs map[string]string) {
	c.Request.Method = "GET"
	u := url.Values{}
	for k, v := range pairs {
		u.Add(k, v)
	}
	c.Request.URL.RawQuery = u.Encode()
}

func SetTestWriteBodyforPost(c *gin.Context, b []byte) {
	c.Request.Method = "POST"
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Body = io.NopCloser(bytes.NewBuffer(b))
}
