package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var testRootDir = filepath.Join(os.Getenv("GO_PROJECT_ROOT"), "apps/funcdiff/test")

func TestFuncDiff(t *testing.T) {
	// pre-step: format go file
	srcPath := filepath.Join(testRootDir, "src1/main.go")
	dstPath := filepath.Join(testRootDir, "src2/main.go")
	for _, name := range []string{
		"fnHello",
		"fnChange",
	} {
		diffEntry, err := funcDiff(srcPath, dstPath, name)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("\ndiff info:")
		fmt.Println(prettySprintFuncInfo(diffEntry.SrcFnInfo))
		fmt.Println("source:\n", diffEntry.SrcFnInfo.Source)
		fmt.Println(prettySprintFuncInfo(diffEntry.DstFnInfo))
		fmt.Println("source:\n", diffEntry.DstFnInfo.Source)
		fmt.Println("diff result:", diffEntry.Result)
	}
}

func TestFuncDiffForGoFiles(t *testing.T) {
	srcPath := filepath.Join(testRootDir, "src1/main.go")
	dstPath := filepath.Join(testRootDir, "src2/main.go")

	diffEntries, err := funcDiffForGoFiles(srcPath, dstPath)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Func diff:")
	for _, entry := range diffEntries {
		fmt.Println(prettySprintDiffEntry(entry) + "\n")
	}
}
