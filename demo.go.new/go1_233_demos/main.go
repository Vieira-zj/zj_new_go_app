package main

import (
	"cmp"
	"log"
)

var version string

func main() {
	log.Printf("version=%s", cmp.Or(version, "dev"))
	log.Println("finish")
}
