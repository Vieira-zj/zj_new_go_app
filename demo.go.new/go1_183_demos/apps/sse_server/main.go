package main

import (
	"flag"
)

// test: http://localhost:8081/

func main() {
	h := flag.Bool("h", false, "help")
	run := flag.String("r", "gin", "run type")

	flag.Parse()

	if *h {
		flag.Usage()
		return
	}

	if *run == "gin" {
		GinSseServe()
	} else {
		SseServe()
	}
}
