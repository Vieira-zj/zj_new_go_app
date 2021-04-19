package pkg

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestSearchProject(t *testing.T) {
	projectID, err := git.SearchProject(context.TODO(), "common-micservice", "microservice")
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
