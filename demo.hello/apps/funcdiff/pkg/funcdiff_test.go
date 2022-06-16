package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var testRootDir = filepath.Join(os.Getenv("GO_PROJECT_ROOT"), "apps/funcdiff/test")

func TestFuncsDiff(t *testing.T) {
	// pre-step: format go file
	srcPath := filepath.Join(testRootDir, "src1/main_format.go")
	dstPath := filepath.Join(testRootDir, "src2/main_format.go")
	for _, name := range []string{
		"fnHello",
		"fnChange",
	} {
		diffEntry, err := FuncDiff(srcPath, dstPath, name)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("\ndiff info:")
		fmt.Println(prettySprintFuncInfo(diffEntry.SrcFnInfo))
		fmt.Println(prettySprintFuncInfo(diffEntry.DstFnInfo))
		fmt.Println("diff result:", diffEntry.Result)
	}
}

func TestGoFileDiffFunc(t *testing.T) {
	// pre-step: format go file
	srcPath := filepath.Join(testRootDir, "src1/main_format.go")
	dstPath := filepath.Join(testRootDir, "src2/main_format.go")

	diffEntries, err := GoFileDiffFunc(srcPath, dstPath)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Func diff:")
	for _, entry := range diffEntries {
		fmt.Println(prettySprintDiffEntry(entry) + "\n")
	}
}
