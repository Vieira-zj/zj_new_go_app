package pkg

import (
	"fmt"
	"testing"
)

func TestGetDiffFilesByCommits(t *testing.T) {
	isDebug = true
	repoPath := "/tmp/test/git_space"

	for _, item := range [][2]string{
		{"ee2b84b71", "693130c64"},
		{"693130c64", "a1765a336"},
		{"a1765a336", "59d198249"},
	} {
		srcHash, dstHash := item[0], item[1]
		diffs, err := getDiffFilesByCommits(repoPath, "master", srcHash, dstHash)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("\ndiff files:")
		for _, diff := range diffs {
			fmt.Printf("diff: %+v\n", diff)
		}
	}
}
