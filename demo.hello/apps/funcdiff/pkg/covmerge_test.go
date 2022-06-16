package pkg

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestPrint(t *testing.T) {
	msg := "hello"
	num := 101
	char := 'a'
	fmt.Printf("print: msg=%q, num=%q, ch=%q\n", msg, num, char)
}

func TestParseProfileLine(t *testing.T) {
	line := "demo.hello/echoserver/handlers/hooks.go:12.65,13.36 1 0"
	fname, profile, err := parseProfileLine(line)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("porfile: [%s]:%+v\n", fname, profile)
}

func TestParseCovFile(t *testing.T) {
	fpath := "/tmp/test/profile.cov"
	res, err := parseCovFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	for fname, profile := range res {
		fmt.Printf("\nmode=%s, filename=%s\n", profile.Mode, fname)
		for _, block := range profile.Blocks {
			fmt.Printf("\tblock: %+v\n", block)
		}
	}
}

func TestParseCovFileForMerge(t *testing.T) {
	fpath := "/tmp/test/profile_merge.cov"
	res, err := parseCovFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	for fname, profile := range res {
		fmt.Printf("\nmode=%s, filename=%s\n", profile.Mode, fname)
		for _, block := range profile.Blocks {
			fmt.Printf("\tblock: %+v\n", block)
		}
	}

}

func TestLinkProfileBlocksToFunc(t *testing.T) {
	srcPath := filepath.Join(testRootDir, "src1/main.go")
	funcInfos, err := GetFuncInfos(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("funcs info:")
	for _, info := range funcInfos {
		fmt.Println(info.Path, info.Name)
	}

	fmt.Println("\nprofile:")
	covPath := "/tmp/test/profile.cov"
	fnProfile, err := parseCovFile(covPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, profile := range fnProfile {
		fmt.Println(profile.FileName)
		for _, block := range profile.Blocks {
			fmt.Printf("\t%+v\n", block)
		}
	}

	fmt.Println("\nlink profile block to func:")
	for _, fnInfo := range funcInfos {
		if profile, ok := fnProfile[fnInfo.Path]; ok {
			funcCoverEntry := linkProfileBlocksToFunc(profile, fnInfo)
			fmt.Println("function:", funcCoverEntry.FuncInfo.Path, funcCoverEntry.FuncInfo.Name)
			fmt.Println("blocks:")
			for _, block := range funcCoverEntry.ProfileBlocks {
				fmt.Printf("\t%+v\n", block)
			}
		}
	}
}
