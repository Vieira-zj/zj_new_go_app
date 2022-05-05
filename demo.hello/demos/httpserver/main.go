package main

import (
	"encoding/json"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"runtime"
	"time"
)

const (
	cntKey     = "Count"
	listenPort = ":8080"
)

func init() {
	// expvar 1. 公共变量 2. 操作都是协程安全的
	count := expvar.NewInt(cntKey)
	count.Set(0)
}

func main() {
	go func() {
		t := time.Tick(time.Second)
		for {
			select {
			case <-t:
				log.Printf("number of goroutines: %d", runtime.NumGoroutine())
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/exec", execHandler)
	mux.HandleFunc("/debug/vars", debugHandler)

	server := &http.Server{
		Addr:              listenPort,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      8 * time.Second,
		Handler:           logging(mux),
	}

	log.Println("http serve at", listenPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalln("server error:", err)
	}
}

/*
Handlers
*/

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello"))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func execHandler(w http.ResponseWriter, r *http.Request) {
	// http server starts 2 goroutines for each request (connection)
	time.Sleep(3 * time.Second)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("exec done"))
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	vars := make([]string, 0, 4)
	expvar.Do(func(kv expvar.KeyValue) {
		if kv.Key == "memstats" || kv.Key == "cmdline" {
			return
		}
		val := fmt.Sprintf("key=%s,value=%s", kv.Key, kv.Value)
		vars = append(vars, val)
	})

	resBody := struct {
		Message string   `json:"message"`
		Expvar  []string `json:"expvars"`
	}{
		Message: "ok",
		Expvar:  vars,
	}
	if err := json.NewEncoder(w).Encode(resBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintln("json encode error:", err)))
	}
}

/*
Middlewares
*/

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/debug/vars" {
			next.ServeHTTP(w, r)
			return
		}

		count := expvar.Get(cntKey).(*expvar.Int)
		count.Add(1)
		log.Printf("access count: %d", count.Value())

		b, err := httputil.DumpRequest(r, false)
		if err != nil {
			log.Println("dump request error:", err)
		}
		log.Println("receive request:\n", string(b))

		next.ServeHTTP(w, r)
	})
}
