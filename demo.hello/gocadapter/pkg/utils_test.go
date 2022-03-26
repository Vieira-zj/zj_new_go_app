package pkg

import (
	"fmt"
	"testing"
)

func TestGetSimpleNowDatetime(t *testing.T) {
	fmt.Println("now:", getSimpleNowDatetime())
}

func TestGetFileNameWithoutExt(t *testing.T) {
	for _, fileName := range []string{"test.json", "sh_output.txt", "results"} {
		fmt.Println("name:", getFileNameWithoutExt(fileName))
	}
}

func TestFormatIPAddress(t *testing.T) {
	addr := "http://127.0.0.1:49970"
	ip, err := formatIPAddress(addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("ip:", ip)
}
