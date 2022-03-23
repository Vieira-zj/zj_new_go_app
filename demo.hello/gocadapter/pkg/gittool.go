package pkg

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

//
// Refer: https://github.com/go-git/go-git/tree/master/_examples
//

// GitClone .
func GitClone(ctx context.Context, URL, workingDir string) (string, error) {
	repo, err := git.PlainCloneContext(ctx, workingDir, false, &git.CloneOptions{
		URL: URL,
		Auth: &http.BasicAuth{
			Username: getParamFromEnv("GITLAB_USER"),
			Password: getParamFromEnv("GITLAB_TOKEN"),
		},
		Progress: os.Stdout,
	})
	if err != nil {
		if errors.Is(err, git.ErrRepositoryAlreadyExists) {
			repoName, err := getRepoNameFromURL(URL)
			if err != nil {
				return "", fmt.Errorf("GitClone error: %w", err)
			}
			log.Printf("Clone repo [%s]: %s", repoName, err.Error())

			r := NewGitRepo(workingDir)
			head, err := r.getRepoHeadCommitShortID()
			if err != nil {
				return "", fmt.Errorf("GitClone error: %w", err)
			}
			return head, nil
		}
		return "", fmt.Errorf("GitClone clone repo error: %w", err)
	}

	ref, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("GitClone get repo head commit error: %w", err)
	}
	return ref.Hash().String()[:8], nil
}

var (
	gitRepo  *GitRepo
	repoOnce sync.Once
)

// GitRepo .
type GitRepo struct {
	name string
	repo *git.Repository
}

// NewGitRepo .
func NewGitRepo(repoPath string) *GitRepo {
	repoOnce.Do(func() {
		repo, err := git.PlainOpen(repoPath)
		if err != nil {
			panic(err)
		}
		gitRepo = &GitRepo{
			name: filepath.Base(repoPath),
			repo: repo,
		}
	})
	return gitRepo
}

// Fetch .
func (r *GitRepo) Fetch(ctx context.Context) error {
	if err := r.repo.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: getParamFromEnv("GITLAB_USER"),
			Password: getParamFromEnv("GITLAB_TOKEN"),
		},
		Progress: os.Stdout,
	}); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			log.Printf("Git fetch [%s]: %s", r.name, err.Error())
			return nil
		}
		return fmt.Errorf("Fetch fetch repo error: %w", err)
	}
	return nil
}

// Pull fetch changes from remote and fast-forward to current branch.
func (r *GitRepo) Pull(ctx context.Context, branch string) (string, error) {
	commitID, err := r.CheckoutBranch(branch)
	if err != nil {
		return "", fmt.Errorf("Pull error: %w", err)
	}

	branchName, err := r.GetBranchFullName(branch)
	if err != nil {
		return "", fmt.Errorf("Pull error: %w", err)
	}

	w, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("Pull get worktree error: %w", err)
	}

	if err = w.PullContext(ctx, &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.ReferenceName(branchName),
		Auth: &http.BasicAuth{
			Username: getParamFromEnv("GITLAB_USER"),
			Password: getParamFromEnv("GITLAB_TOKEN"),
		},
		Progress: os.Stdout,
	}); err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			log.Printf("Pull branch [%s]: %s", branch, err.Error())
			return commitID, nil
		}
		return "", fmt.Errorf("Pull pull branch [%s] error: %w", branch, err)
	}

	head, err := r.getRepoHeadCommitShortID()
	if err != nil {
		return "", fmt.Errorf("Pull error: %w", err)
	}
	return head, nil
}

// CheckoutToCommit .
func (r *GitRepo) CheckoutToCommit(commitID string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("CheckoutToCommit get worktree error: %w", err)
	}

	if err := w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commitID),
	}); err != nil {
		return fmt.Errorf("CheckoutToCommit checkout to commit error: %w", err)
	}
	return nil
}

// CheckoutBranch .
func (r *GitRepo) CheckoutBranch(branch string) (string, error) {
	commitID, err := r.GetBranchCommit(branch)
	if err != nil {
		return "", fmt.Errorf("CheckoutBranch error: %w", err)
	}
	head, err := r.getRepoHeadCommitShortID()
	if err != nil {
		return "", fmt.Errorf("CheckoutBranch error: %w", err)
	}
	if commitID[:8] == head {
		log.Println("Already on branch:", branch)
		return head, nil
	}

	w, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("CheckoutBranch get worktree error: %w", err)
	}

	brName, err := r.GetBranchFullName(branch)
	if err != nil {
		return "", fmt.Errorf("CheckoutBranch error: %w", err)
	}

	if err := w.Checkout((&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(brName),
	})); err != nil {
		return "", fmt.Errorf("CheckoutBranch checkout branch error: %w", err)
	}

	head, err = r.getRepoHeadCommitShortID()
	if err != nil {
		return "", fmt.Errorf("CheckoutBranch error: %w", err)
	}
	return head, nil
}

// CheckoutRemoteBranch creates a new branch to track remote branch.
func (r *GitRepo) CheckoutRemoteBranch(ctx context.Context, branch string) (string, error) {
	if err := r.Fetch(ctx); err != nil {
		return "", fmt.Errorf("CheckoutRemoteBranch error: %w", err)
	}

	_, commitID, err := r.getRemoteBranch(branch)
	if err != nil {
		return "", fmt.Errorf("CheckoutRemoteBranch error: %w", err)
	}

	newBranch := fmt.Sprintf("refs/heads/%s", branch)
	newRef := plumbing.ReferenceName(newBranch)
	if !newRef.IsBranch() {
		return "", fmt.Errorf("CheckoutRemoteBranch invalid ref: %s", newRef.String())
	}
	ref := plumbing.NewHashReference(newRef, plumbing.NewHash(commitID))
	if err = r.repo.Storer.SetReference(ref); err != nil {
		return "", fmt.Errorf("CheckoutRemoteBranch create new branch error: %w", err)
	}
	return commitID[:8], nil
}

// IsBranchExist .
func (r *GitRepo) IsBranchExist(branch string) (bool, error) {
	name, _, err := r.getBranch(branch)
	if err != nil {
		if errors.Is(err, git.ErrBranchNotFound) {
			return false, nil
		}
		return false, err
	}

	if len(name) == 0 {
		return false, nil
	}
	return true, nil
}

// GetBranchFullName .
func (r *GitRepo) GetBranchFullName(branch string) (string, error) {
	name, _, err := r.getBranch(branch)
	return name, err
}

// GetBranchCommit .
func (r *GitRepo) GetBranchCommit(branch string) (string, error) {
	_, commitID, err := r.getBranch(branch)
	return commitID, err
}

func (r *GitRepo) getBranch(branch string) (string, string, error) {
	refs, err := r.repo.References()
	if err != nil {
		return "", "", fmt.Errorf("getBranch get repo refs error: %w", err)
	}

	name := ""
	commitID := ""
	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() && ref.Name().Short() == branch {
			name = ref.Name().String()
			commitID = ref.Hash().String()
		}
		return nil
	})

	if name == "" {
		err = git.ErrBranchNotFound
	}
	return name, commitID, nil
}

// GetRemoteBranchFullName .
func (r *GitRepo) GetRemoteBranchFullName(branch string) (string, error) {
	name, _, err := r.getRemoteBranch(branch)
	return name, err
}

func (r *GitRepo) getRemoteBranch(branch string) (string, string, error) {
	refs, err := r.repo.References()
	if err != nil {
		return "", "", fmt.Errorf("getRemoteBranch get repo refs error: %w", err)
	}

	if !strings.HasPrefix(branch, "origin") {
		branch = fmt.Sprintf("origin/%s", branch)
	}

	name := ""
	commitID := ""
	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsRemote() && ref.Name().Short() == branch {
			name = ref.Name().String()
			commitID = ref.Hash().String()
		}
		return nil
	})

	if name == "" {
		err = git.ErrBranchNotFound
	}
	return name, commitID, err
}

func (r *GitRepo) getRepoHeadCommitShortID() (string, error) {
	ref, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("getRepoHeadCommitShortID get repo head ref error: %w", err)
	}
	return ref.Hash().String()[:8], nil
}
