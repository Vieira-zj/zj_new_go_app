package cover_test

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"demo.apps/go.test/cover"
	"demo.apps/utils"
)

//
// Demo: run integration api test, and do coverage.
//

func TestMain(m *testing.M) {
	if env := os.Getenv("ENV"); env == "dev" {
		log.Println("start http server before test")
		router := cover.InitRouter()
		go cover.HttpServe(router)
	}

	os.Exit(m.Run())
}

func TestAPIPing(t *testing.T) {
	r := utils.NewDefaultHttpRequester()
	url := "http://localhost:8080/ping"
	_, body, err := r.Get(context.Background(), url, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("resp:", string(body))
}

func TestAPIHealthz(t *testing.T) {
	r := utils.NewDefaultHttpRequester()
	url := "http://localhost:8080/healthz"
	_, body, err := r.Get(context.Background(), url, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("resp:", string(body))
}

func TestAPINotFound(t *testing.T) {
	if env := os.Getenv("ENV"); env != "dev" {
		log.Println("skip case TestAPINotFound")
		t.Skip("case TestAPINotFound only run for dev")
	}

	mux := cover.InitRouter()
	r := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("return code is not 404")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("resp body:", string(body))
}
