package utils

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func GoVersion() string {
	return runtime.Version()
}

func GetParentProcessId() int {
	return syscall.Getppid()
}

func KillProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

func GetProjectRootPath() string {
	_, fpath, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(fpath))
}

func GetFnFullName(fn any) string {
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}

func GetCallerInfo(depth int) string {
	pc, _, _, _ := runtime.Caller(depth)
	details := runtime.FuncForPC(pc)
	fpath, line := details.FileLine(pc)

	fullFnName := details.Name()
	idx := strings.LastIndex(fullFnName, ".")
	pkg := fullFnName[:idx]
	fnName := fullFnName[idx+1:]
	return fmt.Sprintf("%s:%d|%s|%s", fpath, line, pkg, fnName)
}

func GetGoroutineID() (int, error) {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// log.Println("read stack size:", n)
	stk := strings.TrimPrefix(string(buf[:n]), "goroutine")

	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		return -1, fmt.Errorf("cannot get goroutine id: %v", err)
	}

	return id, nil
}

var (
	KB = int64(1 << 10)
	MB = int64(1 << 20)
	GB = int64(1 << 30)
)

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("\tAlloc = %v MiB\n", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB\n", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB\n", bToMb(m.Sys))
	fmt.Printf("\tHeapAlloc = %v MiB\n", bToMb(m.HeapAlloc))
	fmt.Printf("\tHeapSys = %v MiB\n", bToMb(m.HeapSys))
	fmt.Printf("\tHeapInuse = %v MiB\n", bToMb(m.HeapInuse))
	fmt.Println()
}

func bToMb(b uint64) uint64 {
	return b / uint64(MB)
}

// refer: https://pkg.go.dev/runtime/debug#SetMemoryLimit
func SetMemoryLimit(size, unit int64) {
	pre := debug.SetMemoryLimit(size * unit)
	log.Printf("mem limit: pre=%dMB, current=%dMB", pre/MB, size)
}

func ForceFreeMemory() {
	debug.FreeOSMemory()
}

func CollectGCStateLoop(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	state := debug.GCStats{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			debug.ReadGCStats(&state)
			log.Printf("%+v", state)
		}
	}
}
