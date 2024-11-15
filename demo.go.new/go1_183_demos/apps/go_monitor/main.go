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
	"syscall"
	"time"

	"github.com/arl/statsviz"
	"github.com/shirou/gopsutil/process"
)

func main() {
	fmt.Println("monitor app start")
	statsviz.RegisterDefault()

	ctx, stop := signal.NotifyContext(context.TODO(), syscall.SIGTERM, syscall.SIGINT)

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
				fmt.Printf("process cpu usage: %.2f, mem usage: %.2f, time: %s\n", cpuPercent, memPercent, time.Now().Format(time.DateOnly))
			}
			fmt.Println("--------------- div ------------------")
		}
		time.Sleep(5 * time.Second)
	}
}

//nolint:unused
func printUsageV2() {
	pid := os.Getegid()
	fmt.Println("current process pid:", pid)

	p, err := process.NewProcess(int32(pid))
	if err != nil {
		panic("new process error:" + err.Error())
	}

	name, err := p.Name()
	if err != nil {
		panic("get process name error:" + err.Error())
	}
	fmt.Println("current process name:", name)

	for {
		if cpuPercent, err := p.CPUPercent(); err != nil {
			fmt.Println("get cpu per error:", err)
		} else {
			fmt.Printf("process cpu usage: %.2f, time: %v\n", cpuPercent, time.Now().Format(time.DateTime))
		}

		if cpuTime, err := p.Times(); err != nil {
			fmt.Println("get cpu time error:", err)
		} else {
			fmt.Println("process cpu time:", cpuTime.String())
			// {"cpu":"cpu","user":0.1,"system":0.1,"idle":0.0,"nice":0.0,"iowait":0.0,"irq":0.0,"softirq":0.0,"steal":0.0,"guest":0.0,"guestNice":0.0}
			// user 和 system 表示的是 CPU 真正执行的用户程序和内核程序的 CPU 使用率，nice 表示的是优先级较低的用户程序的 CPU 使用率，这 3 个指标通常被称为 CPU 的总体使用率
		}

		if memPercent, err := p.MemoryPercent(); err != nil {
			fmt.Println("get memory per error:", err)
		} else {
			fmt.Printf("process memory usage: %.2f", memPercent)
		}

		if memInfo, err := p.MemoryInfo(); err != nil {
			fmt.Println("get memory info error:", err)
		} else {
			fmt.Println("process mem info:", memInfo.String())
			// {"rss":74186752,"vms":419716349952,"hwm":0,"data":0,"stack":0,"locked":0,"swap":0}
			// rss (Resident Set Size): 进程使用的物理内存大小，包含共享库占用的内存
			// vms (Virtual Memory Size): 进程使用的虚拟内存大小，包括堆、栈、共享库和映射文件占用的内存
		}

		fmt.Println("--------------- div ------------------")
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
