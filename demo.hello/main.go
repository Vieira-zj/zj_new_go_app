package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"

	"demo.hello/demos"
)

var (
	version   string
	buildTime string
	goVersion string
)

func init() {
	// go run -ldflags "-X main.version=1.0.0 -X 'main.buildTime=`date`' -X 'main.goVersion=`go version`'" main.go
	if len(version) > 0 {
		fmt.Printf("Run info:\n%s\n%s\n%s\n", version, buildTime, goVersion)
	}

	// for io process, default GOMAXPROCS is min, and prefer to set as "5 * NumCPU"
	fmt.Println("\nApp info:")
	fmt.Printf("cpu threads count: %d\n", runtime.NumCPU())
	fmt.Printf("os threads count: %d\n", runtime.GOMAXPROCS(-1))
	fmt.Printf("goroutines count: %d\n\n", runtime.NumGoroutine())
}

func main() {
	flag.Parse()
	inputArgs := flag.Args()
	if len(inputArgs) > 0 {
		fmt.Println("input args:", strings.Join(inputArgs, ","))
	}

	demos.DemoMain()
	fmt.Println("hello golang")
}
