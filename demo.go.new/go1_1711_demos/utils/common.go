package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// GetCallerInfo: Caller 函数会报告当前 Go 程序调用栈所执行的函数的文件和行号信息。参数 skip 为要上溯的栈帧数。
func GetCallerInfo(skip int) (string, error) {
	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {
		return "", fmt.Errorf("runtime.Caller() failed")
	}

	funcName := runtime.FuncForPC(pc).Name()
	funcName = filepath.Base(funcName)
	return fmt.Sprintf("fnname:%s, file:%s, line:%d", funcName, file, lineNo), nil
}
