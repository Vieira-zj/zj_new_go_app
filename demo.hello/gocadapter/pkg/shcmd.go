package pkg

import (
	"sync"

	"demo.hello/utils"
)

// ShCmd .
type ShCmd struct {
	sh string
}

var (
	shCmd     *ShCmd
	shCmdOnce sync.Once
)

// NewShCmd .
func NewShCmd() *ShCmd {
	shCmdOnce.Do(func() {
		shCmd = &ShCmd{
			sh: utils.GetShellPath(),
		}
	})
	return shCmd
}

// CloneProject .
func (c *ShCmd) CloneProject(uri, moduleDir string) error {
	return nil
}

// GoToolCreateCoverFuncReport .
func (c *ShCmd) GoToolCreateCoverFuncReport(workingPath string) error {
	// cd ${project_root}; go tool cover -func=${input.cov} -o ${output.txt}
	return nil
}

// GoToolCreateCoverHTMLReport .
func (c *ShCmd) GoToolCreateCoverHTMLReport(workingPath string) error {
	// cd ${project_root}; go tool cover -html=${input.cov} -o ${output.html}
	return nil
}
