package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"demo.hello/demos"
)

var (
	version   string
	buildTime string
	goVersion string
)

func httpServe() {
	server := http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second)
		fmt.Fprint(w, "hello world")
	})

	fmt.Println("http serve at :8080")
	go server.ListenAndServe()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	<-ctx.Done()

	stop() // 重置 os.Interrupt 的默认行为
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	if err := server.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}
}

func main() {
	isServe := flag.Bool("s", false, "run http server")
	help := flag.Bool("h", false, "help")
	flag.Parse()

	if *help {
		flag.Usage()

		// go run -ldflags "-X main.version=1.0.0 -X 'main.buildTime=`date`' -X 'main.goVersion=`go version`'" main.go -h
		if len(version) > 0 {
			fmt.Printf("\nBuild info:\n%s\n%s\n%s\n", version, buildTime, goVersion)
		}

		// for io process, default GOMAXPROCS is min, and prefer to set as "5 * NumCPU"
		fmt.Println("\nApp info:")
		fmt.Printf("cpu threads count: %d\n", runtime.NumCPU())
		fmt.Printf("os threads count: %d\n", runtime.GOMAXPROCS(-1))
		fmt.Printf("goroutines count: %d\n\n", runtime.NumGoroutine())
		return
	}

	inputArgs := flag.Args()
	if len(inputArgs) > 0 {
		fmt.Println("input args:", strings.Join(inputArgs, ","))
	}

	if *isServe {
		httpServe()
	} else {
		demos.DemoMain()
	}
	fmt.Println("hello golang")
}
