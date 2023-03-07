package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"unsafe"
)

// GetExecParentPath returns parent dir path for current exec bin file.
func GetExecParentPath() (string, error) {
	fpath, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}

	fAbsPath, err := filepath.Abs(fpath)
	if err != nil {
		return "", err
	}

	return path.Dir(fAbsPath), nil
}

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

func Signature(secret, data []byte) (string, error) {
	hash := hmac.New(sha256.New, secret)
	if _, err := hash.Write(data); err != nil {
		return "", err
	}

	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return signature, nil
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

// unsafe: 使用 unsafe 进行 string和[]byte 转换，避开内存 copy 来提高性能

func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
