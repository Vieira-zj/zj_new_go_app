package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBindReqQuery(t *testing.T) {
	type testQuery struct {
		VersionID string `form:"version_id"`
		Count     uint8  `form:"count"`
		Success   bool   `form:"success"`
	}

	w := httptest.NewRecorder()
	c := GetTestGinContext(w)
	SetTestUrlQueryForGet(c, map[string]string{
		"version_id": "v1.1",
		"count":      "19",
		"success":    "true",
	})
	t.Log("request url:", c.Request.URL.String())

	query := testQuery{}
	err := c.ShouldBindQuery(&query)
	assert.NoError(t, err)
	t.Logf("bind request: %+v", query)
}

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
		URL: &url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
		},
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
