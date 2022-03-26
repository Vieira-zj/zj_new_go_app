package pkg

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

//
// run: go test -timeout 10s -run ^TestCheckoutToCommit$ demo.hello/gocadapter/pkg -v -count=1
//

var (
	testRepoPath = "/tmp/test/echoserver"
)

func TestGetLastDirName(t *testing.T) {
	fmt.Println(filepath.Base(testRepoPath))
}

func TestGitClone(t *testing.T) {
	url := testGetRepoURL()
	fmt.Println("repo url:", url)
	commitID, err := GitClone(context.Background(), url, testRepoPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("repo head:", commitID)
}

func TestFetch(t *testing.T) {
	repo := NewGitRepo(testRepoPath)
	if err := repo.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestPull(t *testing.T) {
	repo := NewGitRepo(testRepoPath)
	head, err := repo.Pull(context.Background(), "rm_staging_copied")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("pull:", head)
}

func TestGetRepoHeadCommitID(t *testing.T) {
	repo := NewGitRepo(testRepoPath)
	head, err := repo.getRepoHeadCommitShortID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("head:", head)
}

func TestGetBranchFullName(t *testing.T) {
	repo := NewGitRepo(testRepoPath)
	for _, branch := range []string{"master"} {
		name, err := repo.GetBranchFullName(branch)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("branch:", name)
	}
}

func TestGetBranchCommit(t *testing.T) {
	branch := "rm_staging_copied"
	repo := NewGitRepo(testRepoPath)
	commitID, err := repo.GetBranchCommit(branch)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("branch [%s] commit: %s\n", branch, commitID[:8])
}

func TestGetRemoteBranchFullName(t *testing.T) {
	repo := NewGitRepo(testRepoPath)
	for _, branch := range []string{"release", "rm_staging_test"} {
		name, err := repo.GetRemoteBranchFullName(branch)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("remote branch:", name)
	}
}

func TestCheckoutToCommit(t *testing.T) {
	repo := NewGitRepo(testRepoPath)
	if err := repo.CheckoutToCommit("a6023e5e"); err != nil {
		t.Fatal(err)
	}
}

func TestCheckoutBranch(t *testing.T) {
	branch := "rm_staging_copied"
	repo := NewGitRepo(testRepoPath)
	head, err := repo.CheckoutBranch(branch)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("checkout branch [%s]: %s\n", branch, head)
}

func TestCheckoutRemoteBranch(t *testing.T) {
	repo := NewGitRepo(testRepoPath)
	commitID, err := repo.CheckoutRemoteBranch(context.Background(), "rm_staging_test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("checkout remote branch:", commitID)
}

func testGetRepoURL() string {
	if err := mockLoadConfig("/tmp/test"); err != nil {
		panic(err)
	}

	srvName := "echoserver"
	val, ok := ModuleToRepoMap[srvName]
	if !ok {
		panic(fmt.Sprintf("service [%s] is not found", srvName))
	}
	return val
}
