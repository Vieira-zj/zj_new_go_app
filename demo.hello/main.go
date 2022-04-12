package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"demo.hello/demos"
	"demo.hello/utils"
	"github.com/fatih/color"
)

var (
	version   string
	buildTime string
	goVersion string
)

func colorPrint() {
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

	fmt.Println()
	logger := utils.NewSimpleLog()
	logger.Debug("debug test, and output to file.")
	logger.Info("info test, and output to file.")
	logger.Warning("warnning test.")
	logger.Error("error test.")
	logger.Fatal("fatal test.")
}

func httpServe() {
	addr := ":8080"
	server := &http.Server{
		Addr: addr,
	}

	// curl http://localhost:8080/
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		fmt.Fprint(w, "hello world")
	})

	fmt.Printf("http serve at %s\n", addr)
	go func() {
		runtime.LockOSThread()
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("serve error:", err)
		}
	}()

	// ctrl-c: signal interrupt
	// ctrl-d: io eof
	//
	// kill (no param) default send syscall.SIGTERM (terminate)
	// kill -2 is syscall.SIGINT (interrupt)
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	<-ctx.Done()

	stop() // 重置 os.Interrupt 的默认行为
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	if err := server.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}
}

func goRoutinesRunOrder() {
	runtime.GOMAXPROCS(1)
	for i := 0; i < 6; i++ {
		go func(i int) {
			fmt.Println(i)
		}(i)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	<-ctx.Done()
}

func main() {
	help := flag.Bool("h", false, "help")
	isColor := flag.Bool("c", false, "run color print demo.")
	isServe := flag.Bool("s", false, "run http server.")
	isGpm := flag.Bool("g", false, "run goroutines with GOMAXPROCS=1.")
	flag.Parse()

	fullPathApp := os.Args[0]
	app := filepath.Base(fullPathApp)
	fmt.Printf("run app (%s): %s\n", fullPathApp, app)

	inputArgs := flag.Args()
	if len(inputArgs) > 0 {
		fmt.Println("input args:", strings.Join(inputArgs, ","))
	}

	if *help {
		flag.Usage()

		// build:
		// go build -ldflags "-X main.version=1.0.0 -X 'main.buildTime=`date`' -X 'main.goVersion=`go version`'" main.go
		// build and run:
		// go run -ldflags "-X main.version=1.0.0 -X 'main.buildTime=`date`' -X 'main.goVersion=`go version`'" main.go -h
		if len(version) > 0 {
			fmt.Printf("\nBuild info:\nversion: %s\n", version)
			fmt.Printf("build time: %s\n", buildTime)
			fmt.Println(goVersion)
		}

		// for io process, default GOMAXPROCS is min, and prefer to set as "N * NumCPU"
		fmt.Println("\nApp info:")
		fmt.Printf("cpu threads count: %d\n", runtime.NumCPU())
		fmt.Printf("os threads count: %d\n", runtime.GOMAXPROCS(-1))
		fmt.Printf("goroutines count: %d\n", runtime.NumGoroutine())

		fmt.Println("\nDeps info:")
		if info, ok := debug.ReadBuildInfo(); ok {
			for _, dep := range info.Deps {
				fmt.Printf("%s=%s\n", dep.Path, dep.Version)
			}
		}
		return
	}

	if *isColor {
		colorPrint()
		return
	}
	if *isServe {
		httpServe()
		return
	}
	if *isGpm {
		goRoutinesRunOrder()
		return
	}

	demos.Main()
	os.Exit(0)
}
