package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestGetHello(t *testing.T) {
	expect := "Hello, World!"
	ret := getHello()
	if ret != expect {
		t.Errorf("expect %s and actual %s.", ret, expect)
	}
}

func TestIndexHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	c := echo.New().NewContext(req, w)
	if err = IndexHandler(c); err != nil {
		t.Fatal(err)
	}

	expect := "Hello, World!"
	ret := string(w.Body.Bytes())
	if ret != expect {
		t.Errorf("expect %s and actual %s.", expect, ret)
	}
}
