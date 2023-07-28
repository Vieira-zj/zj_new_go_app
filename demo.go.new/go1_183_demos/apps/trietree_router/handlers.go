package main

import (
	"compress/gzip"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func defaultHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "method=%s, path=%s", req.Method, req.URL.Path)
}

func testJsonHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"code":0,"msg":"success"}`))
}

func testGzipHandler(w http.ResponseWriter, req *http.Request) {
	// log.Printf("req headers: %+v", req.Header)
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("this is plain text test"))
		return
	}

	fpath := "/tmp/test/raw.json"
	b, err := os.ReadFile(fpath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("open file error: " + err.Error()))
		return
	}

	gzipw := gzip.NewWriter(w)
	if _, err = gzipw.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("gzip compress error: " + err.Error()))
		return
	}
	defer func() {
		if err = gzipw.Flush(); err != nil {
			log.Println("gzip writer flush error:", err)
		}
		if err = gzipw.Close(); err != nil {
			log.Println("gzip writer close error:", err)
		}
	}()

	w.Header().Set("Content-Encoding", "gzip")
	// http: superfluous response.WriteHeader call from main.TestGzipHandler (handlers.go:49)
	// w.WriteHeader(http.StatusOK)
}
