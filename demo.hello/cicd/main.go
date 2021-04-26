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
	for _, key := range keys {
		tree.SubmitIssue(key)
	}
	tree.WaitDone()
	fmt.Println(pkg.GetIssuesTreeText(tree))
	fmt.Println(pkg.GetIssuesTreeUsageText(tree))
}

func refreshData() {
	// TODO: refresh data with interval
}

func removeExpiredData() {
	now := time.Now().Unix()
	delKeys := make([]string, 0)
	for key, tree := range serve.TreeMap {
		if now > tree.GetExpired() {
			delKeys = append(delKeys, key)
		}
	}
	for _, key := range delKeys {
		tree := serve.TreeMap[key]
		if !tree.IsRunning() {
			delete(serve.TreeMap, key)
			fmt.Printf("Store [%s] is expired and removed.\n", key)
		}
	}
}

func main() {
	flag.BoolVar(&help, "h", false, "help.")
	flag.BoolVar(&server, "server", false, "run server mode.")
	flag.BoolVar(&cli, "cli", false, "run command line mode.")
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

			e.POST("/store/save", serve.StoreIssues)
			e.POST("/store/usage", serve.StoreUsage)

			e.POST("/get/store", serve.GetStoreIssues)
			e.POST("/get/issue", serve.GetSingleIssue)
			e.POST("/get/repos", serve.GetRepos)

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
					refreshData()
					removeExpiredData()
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			panic(err)
		}
	}
	fmt.Println("Server stopped.")
}
