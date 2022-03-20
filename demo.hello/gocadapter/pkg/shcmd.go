package pkg

import (
	"errors"
	"fmt"
	"sync"

	"demo.hello/utils"
)

// ShCmd .
type ShCmd struct {
	sh   string
	root string
}

var (
	shCmd     *ShCmd
	shCmdOnce sync.Once
)

// NewShCmd .
func NewShCmd(root string) *ShCmd {
	shCmdOnce.Do(func() {
		shCmd = &ShCmd{
			sh:   utils.GetShellPath(),
			root: root,
		}
	})
	return shCmd
}

// CloneProject .
func (c *ShCmd) CloneProject(uri string) error {
	if err := utils.MakeDir(c.root); err != nil && !errors.Is(err, utils.ErrDirExist) {
		return fmt.Errorf("CloneProject make dir failed: %w", err)
	}

	cmd := fmt.Sprintf("git clone -q %s", uri)
	if _, err := utils.RunShellCmd(c.sh, "-c", cmd); err != nil {
		return fmt.Errorf("CloneProject git clone failed: %w", err)
	}
	return nil
}

// CheckoutBranchFromRemotes .
func CheckoutBranchFromRemotes(name string) {
	// git checkout -b ${branch} remotes/origin/${branch}
}

// SyncBranchWithRemotes .
func SyncBranchWithRemotes(name string) {
	// git checkout ${branch}
	// git fetch origin ${branch}
	// git rebase origin/${branch}
	//
	// or
	// git checkout ${branch}
	// git pull --rebase origin/${branch}
}

// CheckoutToCommit git checkout to specific commit.
func (c *ShCmd) CheckoutToCommit(commitID string) error {
	cmd := fmt.Sprintf("cd %s; git checkout %s", c.root, commitID)
	if _, err := utils.RunShellCmd(c.sh, "-c", cmd); err != nil {
		return fmt.Errorf("CheckoutToCommit git checkout commit failed: %w", err)
	}
	return nil
}

// GoToolCreateCoverFuncReport .
func (c *ShCmd) GoToolCreateCoverFuncReport() error {
	// cd ${project_root}; go tool cover -func=${input.cov} -o ${output.txt}
	return nil
}

// GoToolCreateCoverHTMLReport .
func (c *ShCmd) GoToolCreateCoverHTMLReport() error {
	// cd ${project_root}; go tool cover -html=${input.cov} -o ${output.html}
	return nil
}
