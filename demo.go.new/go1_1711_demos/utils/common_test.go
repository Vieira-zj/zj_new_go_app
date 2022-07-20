package utils

import (
	"runtime"
	"testing"
)

func TestRuntimeCaller(t *testing.T) {
	_, file, lineNo, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("get caller info failed")
	}
	t.Logf("caller: file=%s, line=%d", file, lineNo)

	_, file, lineNo, ok = runtime.Caller(1)
	if !ok {
		t.Fatal("get parent caller info failed")
	}
	t.Logf("parent caller: file=%s, line=%d", file, lineNo)
}

func TestGetCallerInfo(t *testing.T) {
	for i := 1; i < 3; i++ {
		info, err := GetCallerInfo(i)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("caller [%d] info: %s", i, info)
	}
}
