package pkg

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

//
// run: go test -timeout 10s -run ^TestCheckoutBranch$ demo.hello/gocadapter/pkg -v -count=1
//

var (
	testRepoPath = "/tmp/test/git_repos"
)

func TestGetLastDirName(t *testing.T) {
	fmt.Println(filepath.Base(testRepoPath))
}

func TestGitClone(t *testing.T) {
	dir, err := testGetRepoPath()
	if err != nil {
		t.Fatal(err)
	}

	url := getParamFromEnv("GITLAB_REPO_TEST")
	commitID, err := GitClone(context.Background(), url, dir)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("repo head:", commitID)
}

func TestFetch(t *testing.T) {
	path, err := testGetRepoPath()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(path)
	if err := repo.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestPull(t *testing.T) {
	path, err := testGetRepoPath()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(path)
	head, err := repo.Pull(context.Background(), "rm_staging_copied")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("pull:", head)
}

func TestGetRepoHeadCommitID(t *testing.T) {
	path, err := testGetRepoPath()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(path)
	head, err := repo.getRepoHeadCommitShortID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("head:", head)
}

func TestGetBranchFullName(t *testing.T) {
	path, err := testGetRepoPath()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(path)
	for _, branch := range []string{"master"} {
		name, err := repo.GetBranchFullName(branch)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("branch:", name)
	}
}

func TestGetBranchCommit(t *testing.T) {
	path, err := testGetRepoPath()
	if err != nil {
		t.Fatal(err)
	}

	branch := "rm_staging_copied"
	repo := NewGitRepo(path)
	commitID, err := repo.GetBranchCommit(branch)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("branch [%s] commit: %s\n", branch, commitID[:8])
}

func TestGetRemoteBranchFullName(t *testing.T) {
	path, err := testGetRepoPath()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(path)
	for _, branch := range []string{"release", "rm_staging_test"} {
		name, err := repo.GetRemoteBranchFullName(branch)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("remote branch:", name)
	}
}

func TestCheckoutBranch(t *testing.T) {
	path, err := testGetRepoPath()
	if err != nil {
		t.Fatal(err)
	}

	branch := "rm_staging_copied"
	repo := NewGitRepo(path)
	head, err := repo.CheckoutBranch(branch)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("checkout branch [%s]: %s\n", branch, head)
}

func TestCheckoutRemoteBranch(t *testing.T) {
	path, err := testGetRepoPath()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(path)
	commitID, err := repo.CheckoutRemoteBranch(context.Background(), "rm_staging_copied")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("checkout remote branch:", commitID)
}

func testGetRepoPath() (string, error) {
	url := getParamFromEnv("GITLAB_REPO_TEST")
	repoName, err := getRepoNameFromURL(url)
	if err != nil {
		return "", err
	}

	path := filepath.Join(testRepoPath, repoName)
	return path, nil
}
