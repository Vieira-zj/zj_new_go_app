package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestGetCondition01(t *testing.T) {
	cond1 := true
	cond2 := true
	url := fmt.Sprintf("/cover?cond1=%v&cond2=%v", cond1, cond2)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	c := echo.New().NewContext(req, w)
	if err = CoverHandler(c); err != nil {
		t.Fatal(err)
	}

	expect := []byte("AC")
	ret := w.Body.Bytes()
	if bytes.Compare(ret, expect) != 0 {
		t.Errorf("expect %s and actual %s.", string(expect), string(ret))
	}
}
