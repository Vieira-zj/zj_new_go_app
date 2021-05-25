package pkg

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestNewMergeRequest(t *testing.T) {
	git := NewGitlabTool()
	mr, err := NewMergeRequest(context.TODO(), git, os.Getenv("GITLAB_MR_TEST"))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(mr)
}
