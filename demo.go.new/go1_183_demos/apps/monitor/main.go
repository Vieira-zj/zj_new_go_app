package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"time"

	"demo.apps/utils"
	"github.com/arl/statsviz"
	"github.com/shirou/gopsutil/process"
)

func main() {
	fmt.Println("monitor app start")
	statsviz.RegisterDefault()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	go startPProf()
	go printUsage()
	go printGCStats(ctx)
	go printNumGoroutine(ctx)

	time.Sleep(10 * time.Second)

	fmt.Println("before, total goroutines:", runtime.NumGoroutine())
	for i := 0; i < 1e5; i++ {
		go func() {
			time.Sleep(10 * time.Second)
		}()
	}

	<-ctx.Done()
	stop()

	time.Sleep(100 * time.Millisecond)
	fmt.Println("monitor app exit")
}

func printNumGoroutine(ctx context.Context) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()

	for i := 1; ; i++ {
		select {
		case <-t.C:
			fmt.Printf("after %d sec, total goroutines: %d\n", (i+1)*5, runtime.NumGoroutine())
		case <-ctx.Done():
			fmt.Println("printNumGoroutine exit:", ctx.Err().Error())
			return
		}
	}
}

func printUsage() {
	pid := os.Getpid()
	fmt.Println("current process pid:", pid)

	p, err := process.NewProcess(int32(pid))
	if err != nil {
		panic(err)
	}

	for {
		cpuPercent, err := p.CPUPercent()
		if err != nil {
			fmt.Println("get cpu per error:", err)
		}

		if cpuPercent > 0 {
			memPercent, err := p.MemoryPercent()
			if err != nil {
				fmt.Println("get mem per error:", err)
			} else {
				fmt.Printf("process cpu usage: %.2f, mem usage: %.2f, time: %s\n", cpuPercent, memPercent, utils.FormatDateTime(time.Now()))
			}
			fmt.Println("---------------div------------------")
		}
		time.Sleep(5 * time.Second)
	}
}

func printGCStats(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	s := debug.GCStats{}

	for {
		select {
		case <-t.C:
			debug.ReadGCStats(&s)
			fmt.Printf("gc num %d, last@%v, PauseTotal %v\n", s.NumGC, s.LastGC, s.PauseTotal)
		case <-ctx.Done():
			fmt.Println("printGCStats exit:", ctx.Err().Error())
			return
		}
	}
}

func startPProf() {
	fmt.Println(http.ListenAndServe("127.0.0.1:6060", nil))
}
