package pkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetLastDirName(t *testing.T) {
	repoPath, _, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(filepath.Base(repoPath))
}

// run: go test -timeout 1800s -run ^TestGitClone$ demo.hello/gocplugin/pkg -v -count=1
func TestGitClone(t *testing.T) {
	repoPath, repoURL, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("repo url:", repoURL)
	commitID, err := GitClone(context.Background(), repoURL, repoPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("repo head:", commitID)
}

// run: go test -timeout 30s -run ^TestFetch$ demo.hello/gocplugin/pkg -v -count=1
func TestFetch(t *testing.T) {
	repoPath, _, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(repoPath)
	if err := repo.Fetch(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// run: go test -timeout 30s -run ^TestCheckoutRemoteBranch$ demo.hello/gocplugin/pkg -v -count=1
func TestCheckoutRemoteBranch(t *testing.T) {
	repoPath, _, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(repoPath)
	commitID, err := repo.CheckoutRemoteBranch(context.Background(), "staging_for_cover")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("checkout remote branch:", commitID)
}

// run: go test -timeout 30s -run ^TestPull$ demo.hello/gocplugin/pkg -v -count=1
func TestPull(t *testing.T) {
	// error: ssh: handshake failed
	// fix: git clone repo by https instead of ssh
	repoPath, _, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}

	branch := "staging_for_cover"
	repo := NewGitRepo(repoPath)
	ok, err := repo.IsBranchExist(branch)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("branch [%s] is not found", branch)
	}

	head, err := repo.Pull(context.Background(), branch)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("pull:", head)
}

func TestGetRepoHeadCommitID(t *testing.T) {
	repoPath, _, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(repoPath)
	head, err := repo.getRepoHeadCommitShortID()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("head:", head)
}

func TestGetBranchFullName(t *testing.T) {
	repoPath, _, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(repoPath)
	for _, branch := range []string{"master"} {
		name, err := repo.GetBranchFullName(branch)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("branch:", name)
	}
}

func TestGetBranchCommit(t *testing.T) {
	repoPath, _, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}

	branch := "rm_staging_copied"
	repo := NewGitRepo(repoPath)
	commitID, err := repo.GetBranchCommit(branch)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("branch [%s] commit: %s\n", branch, commitID[:8])
}

func TestGetRemoteBranchFullName(t *testing.T) {
	repoPath, _, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(repoPath)
	for _, branch := range []string{"release", "rm_staging_test"} {
		name, err := repo.GetRemoteBranchFullName(branch)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("remote branch:", name)
	}
}

// run: go test -timeout 10s -run ^TestCheckoutToCommit$ demo.hello/gocplugin/pkg -v -count=1
func TestCheckoutToCommit(t *testing.T) {
	repoPath, _, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(repoPath)
	if err := repo.CheckoutToCommit("8fce725f646"); err != nil {
		t.Fatal(err)
	}
}

// run: go test -timeout 10s -run ^TestCheckoutBranch$ demo.hello/gocplugin/pkg -v -count=1
func TestCheckoutBranch(t *testing.T) {
	repoPath, _, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}

	branch := "staging_for_cover"
	repo := NewGitRepo(repoPath)
	head, err := repo.CheckoutBranch(branch)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("checkout branch [%s]: %s\n", branch, head)
}

func TestIsBranchExist(t *testing.T) {
	repoPath, _, err := testGetRepoPathAndURL()
	if err != nil {
		t.Fatal(err)
	}

	repo := NewGitRepo(repoPath)
	ok, err := repo.IsBranchExist("staging_for_cover")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("branch exist:", ok)
}

func testGetRepoPathAndURL() (string, string, error) {
	root := filepath.Join(os.Getenv("HOME"), "Downloads/data/goc_staging_space")
	if err := InitConfig(root); err != nil {
		return "", "", err
	}

	srvName := "srv_name"
	val, ok := ModuleToRepoMap[srvName]
	if !ok {
		err := fmt.Errorf("service [%s] is not found in map", srvName)
		return "", "", err
	}
	return filepath.Join(root, srvName, "repo"), val, nil
}
