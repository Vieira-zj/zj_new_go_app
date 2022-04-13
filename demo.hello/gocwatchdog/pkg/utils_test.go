package pkg

import (
	"fmt"
	"testing"

	"demo.hello/utils"
)

func TestGetSimpleNowDatetime(t *testing.T) {
	fmt.Println("now:", getSimpleNowDatetime())
}

func TestGetFilePathWithNewExt(t *testing.T) {
	filePath := "/tmp/test/output.txt"
	result := GetFilePathWithNewExt(filePath, "html")
	fmt.Println(result)
}

func TestGetFilePathWithoutExt(t *testing.T) {
	for _, fileName := range []string{"test.json", "sh_output.txt", "results"} {
		fmt.Println("name:", getFilePathWithoutExt(fileName))
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

func TestReadCoverTotalFromResults(t *testing.T) {
	filePath := "/tmp/test/echoserver/cover_data/staging_th_apa_goc_echoserver_master_845820727e_20220331_182224.func"
	lines, err := utils.ReadLinesFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	summary := lines[len(lines)-1]
	fmt.Println("cover total:", getCoverTotalFromSummary(summary))
}
