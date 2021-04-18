package pkg

import (
	"context"
	"fmt"
	"os"
	"testing"
)

var (
	git = NewGitlabTool()
)

func TestSearchProject(t *testing.T) {
	projectID, err := git.SearchProject(context.TODO(), "kyc_service", "base")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("project id:", projectID)
}

func TestGetSingleMR(t *testing.T) {
	mr := os.Getenv("GITLAB_MR_TEST")
	resp, err := git.GetSingleMR(context.TODO(), mr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(resp))
}
