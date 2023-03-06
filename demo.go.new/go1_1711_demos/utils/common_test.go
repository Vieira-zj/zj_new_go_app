package utils

import (
	"runtime"
	"testing"
)

func TestGetExecParentPath(t *testing.T) {
	fPath, err := GetExecParentPath()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("exec parent path:", fPath)
}

func TestSignature(t *testing.T) {
	key := []byte("1f8a1b84ef9b6fb565934715254dc2")
	data := []byte(`{"request_id":"16708278950054680fx","update_time_from":1670827884,"update_time_to":1670827885,"ts":"1670827895017"}`)
	s, err := Signature(key, data)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("signature:", s)
}

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
		t.Logf("caller [%d] info: %+v", i, info)
	}
}

func TestStr2bytes(t *testing.T) {
	bs := Str2bytes("hello")
	t.Logf("size: %d", len(bs))
	for _, b := range bs {
		t.Logf("%c", b)
	}
}

func TestBytes2str(t *testing.T) {
	s := Bytes2str([]byte{'h', 'e', 'l', 'l', 'o'})
	t.Log("string:", s)
}
