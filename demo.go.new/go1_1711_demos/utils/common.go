package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

func JsonDumps(obj interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(obj)
	return buf.String(), err
}

func JsonLoads(value string, obj interface{}) error {
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	return decoder.Decode(obj)
}

type CallerInfo struct {
	FnName string `json:"fn_name"`
	File   string `json:"file"`
	LineNo int    `json:"line_no,string"`
}

// GetCallerInfo: Caller 函数会报告当前 Go 程序调用栈所执行的函数的文件和行号信息。参数 skip 为要上溯的栈帧数。
func GetCallerInfo(skip int) (CallerInfo, error) {
	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {
		return CallerInfo{}, fmt.Errorf("runtime.Caller() failed")
	}

	funcName := runtime.FuncForPC(pc).Name()
	funcName = filepath.Base(funcName)
	return CallerInfo{
		FnName: funcName,
		File:   file,
		LineNo: lineNo,
	}, nil
}
