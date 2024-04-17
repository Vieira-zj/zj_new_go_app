package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func GetParentProcessId() int {
	return syscall.Getppid()
}

func KillProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

func GetFullFnName(fn any) string {
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
	stk := strings.TrimPrefix(string(buf[:n]), "goroutine")

	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		return -1, fmt.Errorf("cannot get goroutine id: %v", err)
	}

	return id, nil
}
