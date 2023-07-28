package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	rtype := flag.String("type", "normal", "run type: normal, trietree")
	flag.Parse()

	if *rtype == "trietree" {
		trietreeHttpServe()
	} else {
		httpServe()
	}
}

// Http Server
//
// http://127.0.0.1:8081/
// curl 127.0.0.1:8081/ping
// curl 127.0.0.1:8081/test/gzip -v -H "Accept-Encoding: gzip" --output data.gzip
//

func httpServe() {
	// Note: url path pattern must be suffix with "/", like "/public/"
	http.Handle("/", http.FileServer(http.Dir("./public")))

	http.HandleFunc("/ping", defaultHandler)
	http.HandleFunc("/test/gzip", testGzipHandler)

	port := ":8081"
	log.Println("http serve at:", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// TrieTree Http Server
//
// curl 127.0.0.1:8081/hello
// curl 127.0.0.1:8081/users/list
// curl 127.0.0.1:8081/notfound
//
// curl 127.0.0.1:8081/test/json
// curl 127.0.0.1:8081/test/gzip -v -H "Accept-Encoding: gzip" --output data.gzip; cat data.gzip | gunzip
//

func trietreeHttpServe() {
	r := NewRouter()
	r.Add("GET", "/hello", defaultHandler)
	r.Add("GET", "/users/list", defaultHandler)

	r.Add("GET", "/test/json", testJsonHandler)
	r.Add("GET", "/test/gzip", testGzipHandler)

	port := ":8081"
	log.Println("trietree http serve at:", port)
	log.Fatal(http.ListenAndServe(port, r))
}

// Router by TrieTree

type Router struct {
	method   string
	path     string
	isPath   bool
	children map[byte]*Router
	handler  http.HandlerFunc
}

func NewRouter() *Router {
	return &Router{isPath: false, children: map[byte]*Router{}}
}

func (r *Router) Add(method, path string, handler http.HandlerFunc) {
	parent := r
	for i := range path {
		if child, ok := parent.children[path[i]]; ok {
			parent = child
		} else {
			child := NewRouter()
			parent.children[path[i]] = child
			parent = child
		}
	}

	parent.path = path
	parent.method = method
	parent.isPath = true
	parent.handler = handler
}

func (r *Router) Find(method, path string) (http.HandlerFunc, bool) {
	parent := r
	for i := range path {
		if child, ok := parent.children[path[i]]; ok {
			parent = child
		} else {
			return nil, false
		}
	}

	return parent.handler, parent.isPath && parent.method == method
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if handler, ok := r.Find(req.Method, req.URL.Path); ok {
		handler(w, req)
	} else {
		http.NotFound(w, req)
	}
}
