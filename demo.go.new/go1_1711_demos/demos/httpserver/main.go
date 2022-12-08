package main

import (
	"context"
	"encoding/json"
	"go1_1711_demo/utils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
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
// Mem Profile:
// curl http://localhost:8082/mem/profile
// curl http://localhost:8082/mem/clear
//

type Dest struct {
	Cid        string `json:"cid"`
	InstanceID string `json:"instance_id"`
}

var dests = map[string]string{
	"sg": "ce7980b8a02bc1",
	"id": "662e0e2c5e3c12",
}

var (
	locker = sync.Mutex{}
	data   = make([][]int64, 0)
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go processMoitor(ctx)

	listenPort, ok := os.LookupEnv("PORT")
	if !ok {
		log.Fatalln("get port failed")
	}

	server := &http.Server{
		Addr:              ":" + listenPort,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      8 * time.Second,
		Handler:           initRoute(),
	}

	log.Println("http serve at", listenPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalln("server error:", err)
	}
}

func initRoute() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/dest", setDestHandler)

	mux.HandleFunc("/mem/profile", memProfileHandler)
	mux.HandleFunc("/mem/clear", memClearHandler)
	return mux
}

// Handlers

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

func memProfileHandler(w http.ResponseWriter, r *http.Request) {
	locker.Lock()
	defer locker.Unlock()

	s := make([]int64, 1000)
	data = append(data, s)
	time.Sleep(30 * time.Millisecond)
	w.Write([]byte("memory allocated"))
}

func memClearHandler(w http.ResponseWriter, r *http.Request) {
	data = make([][]int64, 0)
	w.Write([]byte("memory clear"))
}

// Process Monitor

func processMoitor(ctx context.Context) {
	monitor, err := utils.NewProcessMonitor()
	if err != nil {
		log.Fatalln("init process monitor error:", err)
	}
	log.Println("start process monitor")

	const interval = time.Duration(5)
	tick := time.Tick(interval * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Println("process monitor exit:", ctx.Err())
		case <-tick:
			func() {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				result, err := monitor.PrettyString(ctx)
				if err != nil {
					log.Println("monitor error:", err)
				}
				log.Println("usage:", result)
			}()
		}
	}
}
