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
	"unsafe"
)

// String

func MultiSplitString(str string, splits []rune) []string {
	keysDict := make(map[rune]struct{}, len(splits))
	for _, key := range splits {
		keysDict[key] = struct{}{}
	}

	results := strings.FieldsFunc(str, func(r rune) bool {
		_, ok := keysDict[r]
		return ok
	})
	return results
}

// MyString: string demo to reuse bytes.
type MyString struct {
	data []byte
}

func NewMyString() MyString {
	return MyString{
		data: make([]byte, 0),
	}
}

func (s *MyString) SetValue(value string) {
	s.data = append(s.data[:0], value...)
}

func (s *MyString) SetValueBytes(value []byte) {
	s.data = append(s.data[:0], value...)
}

func (s MyString) GetValue() string {
	return Bytes2string(s.data)
}

// Bytes2string: unsafe convert bytes to string.
func Bytes2string(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// String2bytes unsafe convert string to bytes.
func String2bytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
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
	f, line := details.FileLine(pc)
	return fmt.Sprintf("%s:%d %s", f, line, details.Name())
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
