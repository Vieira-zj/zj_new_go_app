package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//
// Http Server Demo
//
// Run: PORT=8082 go run main.go
//
// Test:
// curl http://localhost:8082/ping
// curl -XPOST http://localhost:8082/dest -d '{"cid":"sg","instance_id":"ce7980b8a02bc1"}'
//

type Dest struct {
	Cid        string `json:"cid"`
	InstanceID string `json:"instance_id"`
}

var dests = map[string]string{
	"sg": "ce7980b8a02bc1",
	"id": "662e0e2c5e3c12",
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/dest", setDestHandler)

	listenPort, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatalln("get port failed")
	}

	server := &http.Server{
		Addr:              ":" + listenPort,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      8 * time.Second,
		Handler:           mux,
	}

	log.Println("http serve at", listenPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalln("server error:", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("http server demo ok"))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func setDestHandler(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if r.Body != nil {
		r.Body.Close()
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("read body error: " + err.Error()))
		return
	}

	dest := &Dest{}
	if err := json.Unmarshal(body, dest); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unmarshal body error: " + err.Error()))
		return
	}

	dests[dest.Cid] = dest.InstanceID
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("set dest success"))
}
