package main

import (
	"embed"
	"io"
	"log"
)

//go:embed json/*
var staticFS embed.FS

func main() {
	f, err := staticFS.Open("json/raw.json")
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

	log.Println("read json from static:\n", string(b))
}
