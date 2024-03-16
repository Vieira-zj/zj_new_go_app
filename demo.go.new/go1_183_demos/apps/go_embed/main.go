package main

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"
)

//go:embed static/*
var staticFS embed.FS

type StaticFS struct {
	inner embed.FS
}

func (fs StaticFS) Open(name string) (fs.File, error) {
	log.Println("open embed file:", name)
	return fs.inner.Open(path.Join("static", name))
}

func httpServeWithEmbed() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			log.Fatal(err)
		}
	})

	http.Handle("/", http.FileServer(http.FS(StaticFS{staticFS})))

	log.Println("http serve at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func readFromEmbed() {
	f, err := staticFS.Open("static/raw.json")
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("read embed json:\n", string(b))
}

func main() {
	readFromEmbed()

	httpServeWithEmbed()
}
