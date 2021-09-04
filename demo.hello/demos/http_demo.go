package demos

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
)

const (
	addr = "localhost"
	port = 5002
)

var accessCount int

func startHTTPServerV1() error {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux := buildDefaultServeMux()
		mux.ServeHTTP(w, r)
	})

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return err
	}
	if err = http.Serve(listener, handler); err != nil {
		return fmt.Errorf("[svc] failed to start http: %v", err)
	}
	return nil
}

func startHTTPServerV2() error {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux := buildDefaultServeMux()
		mux.ServeHTTP(w, r)
	})
	dumpHandler := buildDumpHandler(handler)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		Handler: dumpHandler,
	}
	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("[svc] failed to start http: %v", err)
	}
	return nil
}

func buildDefaultServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Write([]byte("index page"))
		} else {
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		accessCount++
		if r.Method == "POST" {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "read body failed", http.StatusInternalServerError)
			}
			defer r.Body.Close()
			fmt.Printf("[svc] receive body: %s\n", b)
		}
		w.Write([]byte(fmt.Sprintf("[svc] access count=%d", accessCount)))
	})

	return mux
}

func buildDumpHandler(src http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if req, err := httputil.DumpRequest(r, true); err != nil {
			fmt.Printf("[svc] could not dump request: %v\n", err)
		} else {
			fmt.Printf("[svc] received request:\n%s\n", string(req))
		}
		src.ServeHTTP(w, r)
	})
}
