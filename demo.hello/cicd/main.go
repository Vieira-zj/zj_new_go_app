package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"demo.hello/cicd/pkg"
	serve "demo.hello/cicd/server"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

var (
	jira = pkg.NewJiraTool()

	help         bool
	cli          bool
	server       bool
	releaseCycle string
)

func printReleaseCycleTree(ctx context.Context, relCycle string) {
	jql := fmt.Sprintf(`"Release Cycle" = "%s"`, relCycle)
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		panic(err)
	}

	tree := pkg.NewJiraIssuesTree(ctx, 8)
	tree.Collect()
	for _, key := range keys {
		tree.SubmitIssue(key)
	}

	for tree.QueueSize() > 0 {
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)
	fmt.Println(tree.ToText())
	tree.PrintUsage()
}

func main() {
	flag.BoolVar(&help, "h", false, "help.")
	flag.BoolVar(&cli, "cli", false, "run command line mode.")
	flag.BoolVar(&server, "server", false, "run server mode.")
	flag.StringVar(&releaseCycle, "releaseCycle", "", "Release Cycle for jira issues.")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	// cli
	if cli {
		if len(releaseCycle) > 0 {
			printReleaseCycleTree(mainCtx, releaseCycle)
		}
		return
	}

	// server
	var e *echo.Echo
	if server {
		go func() {
			e = echo.New()

			e.GET("/", serve.Index)
			e.GET("/ping", serve.Ping)

			e.GET("/store", serve.StoreIssues)
			e.GET("/store/usage", serve.StoreUsage)

			e.GET("/get", serve.GetStoreIssues)
			e.GET("/get/issue", serve.GetSingleIssue)

			e.Logger.SetLevel(log.INFO)
			e.Logger.Fatal(e.Start(":8081"))
		}()

		// schedule job
		go func() {
			c := time.Tick(time.Duration(10) * time.Second)
			for {
				select {
				case <-mainCtx.Done():
					fmt.Println("Schedule job exit.")
					return
				case <-c:
					now := time.Now().Unix()
					delKeys := make([]string, 0, 3)
					for key, tree := range serve.TreeMap {
						if now > tree.GetExpired() {
							delKeys = append(delKeys, key)
						} else if !tree.IsRunning() {
							serve.StoreCancelMap[key]()
						}
					}
					for _, key := range delKeys {
						fmt.Printf("Store [%s] is expired and remmove.\n", key)
						serve.StoreCancelMap[key]()
						delete(serve.TreeMap, key)
					}
				}
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// quit
	<-quit
	if server && e != nil {
		mainCancel()
		for _, cancel := range serve.StoreCancelMap {
			cancel()
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			panic(err)
		}
	}
	fmt.Println("Server stopped.")
}
