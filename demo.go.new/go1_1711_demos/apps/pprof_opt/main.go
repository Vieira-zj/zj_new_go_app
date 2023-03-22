package main

import (
	"bytes"
	"io"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"

	"gitlab.com/golang-commonmark/markdown"
)

func render(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	src, err := io.ReadAll(r.Body)
	if r.Body != nil {
		defer r.Body.Close()
	}
	if err != nil {
		log.Printf("error reading body: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	md := markdown.New(
		markdown.XHTMLOutput(true),
		markdown.Typographer(true),
		markdown.Linkify(true),
		markdown.Tables(true),
	)

	var buf bytes.Buffer
	if err := md.Render(&buf, src); err != nil {
		log.Printf("error converting markdown: %v", err)
		http.Error(w, "Malformed markdown", http.StatusBadRequest)
		return
	}

	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("error writing response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func load() {
	rand.Seed(time.Now().Unix())
	n := rand.Intn(100)
	sum := 0
	for i := 0; i < n; i++ {
		for j := 0; j < 10e4; j++ {
			sum += 1
		}
	}
}

// refer: https://mp.weixin.qq.com/s/elsHIqDQ0yABUZXNVpjwMg

func main() {
	http.HandleFunc("/render", render)
	log.Printf("Serving on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
