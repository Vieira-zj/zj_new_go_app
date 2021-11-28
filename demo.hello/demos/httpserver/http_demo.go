package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
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

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		Handler: buildDumpHandler(handler),
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
			if r.Body != nil {
				r.Body.Close()
			}
			if err != nil {
				http.Error(w, "read body failed", http.StatusInternalServerError)
			}
			fmt.Printf("[svc] receive body: %s\n", b)
		}

		w.Header().Add("XTag", "XServer")
		w.WriteHeader(203)
		w.Write([]byte(fmt.Sprintf("[svc] access count=%d", accessCount)))
	})

	return mux
}

func buildDumpHandler(src http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if req, err := httputil.DumpRequest(r, true); err != nil {
			fmt.Printf("[svc] could not dump request: %v\n", err)
		} else {
			fmt.Printf("[svc] received request:\n%s\n", req)
		}

		recorder := httptest.NewRecorder()
		defer func() {
			if resp, err := httputil.DumpResponse(recorder.Result(), true); err != nil {
				fmt.Printf("[svc] could not dump response: %v", err)
			} else {
				fmt.Printf("[svc] sent response:\n%s\n", resp)
			}
			mirrorResponse(recorder, w)
		}()

		src.ServeHTTP(recorder, r)
	})
}

func mirrorResponse(src *httptest.ResponseRecorder, dst http.ResponseWriter) {
	for k, v := range src.Header() {
		dst.Header().Add(k, v[0])
	}
	dst.Write(src.Body.Bytes())
}
