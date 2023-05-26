package e2e_test

import (
	"log"
	"net/http"
	"testing"

	"demo.apps/go.test/cover"
)

// TestMain: wrap http serve main with cover test.
func TestMain(t *testing.T) {
	exit := make(chan struct{})

	mux := cover.InitRouter()
	mux.HandleFunc("/cover", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		exit <- struct{}{}
	})

	go cover.HttpServe(mux)

	<-exit
	log.Println("get stop cover signal, and exit http serve")
}
