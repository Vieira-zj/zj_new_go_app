package utils

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Slice

func DelFirstNItemsOfSlice(s []any /* will change input slice */, n int) ([]any, error) {
	if n >= len(s) {
		return nil, fmt.Errorf("n must be less than length of input slice")
	}

	m := copy(s, s[n:])
	for i := m; i < len(s); i++ {
		s[i] = nil // avoid memory leaks
	}

	s = s[:m] // reset length
	return s, nil
}

// Datetime

const timeLayout = "2006-01-02 15:04:05"

func FormatDateTime(ti time.Time) string {
	return ti.Format(timeLayout)
}

// Network

func GetLocalIPAddr() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "-1:", errors.New("not happen")
}

func GetLocalIPAddrByDial() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return "", err
	}

	if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		idx := strings.Index(addr.String(), ":")
		return addr.String()[:idx], nil
	}
	return "-1", errors.New("not happen")
}

// Runtime

func GetParentProcessId() int {
	return syscall.Getppid()
}

func KillProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
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

// Reflect

func GetFuncDeclare(fn any) (string, error) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return "", fmt.Errorf("must be function, but got: %s", fnType.Kind().String())
	}

	input := make([]string, 0)
	for i := 0; i < fnType.NumIn(); i++ {
		arg := fnType.In(i)
		input = append(input, arg.String())
	}

	output := make([]string, 0)
	for i := 0; i < fnType.NumOut(); i++ {
		result := fnType.Out(i)
		output = append(output, result.String())
	}

	return fmt.Sprintf("input args:%s, output results:%s", strings.Join(input, ","), strings.Join(output, ",")), nil
}
