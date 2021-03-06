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
	"github.com/fatih/color"
)

var (
	version   string
	buildTime string
	goVersion string
)

func colorDemo() {
	fmt.Println("\nStandard colors")
	color.Cyan("Prints text in cyan.")
	color.Blue("Prints %s in blue.", "text")
	color.Red("We have red")
	color.Magenta("And many others ..")

	fmt.Println("\nMix and reuse colors")
	c := color.New(color.FgCyan).Add(color.Underline)
	c.Println("Prints cyan text with an underline.")
	d := color.New(color.FgCyan, color.Bold)
	d.Printf("This prints bold cyan %s\n", "too!.")

	red1 := color.New(color.FgRed)
	boldRed := red1.Add(color.Bold)
	boldRed.Println("This will print text in bold red.")
	whiteBackground := red1.Add(color.BgWhite)
	whiteBackground.Println("Red text with white background.")

	fmt.Println("\nCustom print functions")
	notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
	notice("Don't forget this...")

	fmt.Println("\nInsert into noncolor strings")
	yellow := color.New(color.FgYellow).SprintFunc()
	red2 := color.New(color.FgRed).SprintFunc()
	fmt.Printf("This is a %s and this is %s.\n", yellow("warning"), red2("error"))

	fmt.Println("This", color.RedString("warning"), "should be not neglected.")
	fmt.Printf("%v %v\n", color.GreenString("Info:"), "an important message.")

	fmt.Println("\nPlug into existing code")
	color.Set(color.FgYellow)
	fmt.Println("Existing text will now be in yellow")
	fmt.Printf("This one %s\n", "too")
	color.Unset()
}

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
	isColor := flag.Bool("c", false, "run color demo")
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

	// go run main.go -c
	if *isColor {
		colorDemo()
		return
	}

	if *isServe {
		httpServe()
	} else {
		demos.DemoMain()
	}
}
