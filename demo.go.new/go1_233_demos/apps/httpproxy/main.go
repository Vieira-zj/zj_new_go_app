package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func main() {
	url, err := url.Parse("https://backend:8082")
	if err != nil {
		log.Fatal("Failed to parse URL:", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	proxy.Transport = &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     time.Minute,
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Error occurred while proxying request: %v", err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
	}

	handler := loggingMiddleware(ratelimitMiddleware(proxy))

	server := &http.Server{
		Addr:         ":8081",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("Proxy server started on :8081")
	defer log.Println("Proxy server stopped")
	if err = server.ListenAndServe(); err != nil {
		log.Fatal("Failed to start proxy server:", err)
	}
}

// Middleware functions

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s | %s | %v millsec", r.Method, r.URL.Path, time.Since(start).Milliseconds())
	})
}

var limit = rate.NewLimiter(rate.Limit(10), 2)

func ratelimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limit.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Utils

var bufferPool = &sync.Pool{
	New: func() any {
		return make([]byte, 4096)
	},
}

func CopyResponse(dst io.Writer, src io.Reader) {
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)
	// io.CopyBuffer in Go's io package is a function used to copy data from an io.Reader source to an io.Writer destination, similar to io.Copy.
	// The key difference is that io.CopyBuffer allows the caller to provide a pre-allocated byte slice as a buffer for the copying operation, rather than io Copy allocating a temporary one internally.
	if _, err := io.CopyBuffer(dst, src, buf); err != nil {
		log.Printf("Failed to copy response: %v", err)
	}
}
