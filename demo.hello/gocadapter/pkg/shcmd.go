package pkg

import (
	"fmt"
	"path/filepath"
	"strings"
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

// Run .
func (c *ShCmd) Run(cmd string) (string, error) {
	return utils.RunShellCmd(c.sh, "-c", cmd)
}

// GoToolCreateCoverFuncReport .
func (c *ShCmd) GoToolCreateCoverFuncReport(workingPath, covFile string) error {
	outFile := strings.Replace(covFile, filepath.Ext(covFile), "func", 1)
	cmd := fmt.Sprintf("cd %s; go tool cover -func=%s -o %s", workingPath, covFile, outFile)
	if _, err := utils.RunShellCmd(c.sh, "-c", cmd); err != nil {
		return fmt.Errorf("GoToolCreateCoverFuncReport run command failed: %s", cmd)
	}
	return nil
}

// GoToolCreateCoverHTMLReport .
func (c *ShCmd) GoToolCreateCoverHTMLReport(workingPath, covFile string) error {
	outFile := strings.Replace(covFile, filepath.Ext(covFile), "html", 1)
	cmd := fmt.Sprintf("cd %s; go tool cover -html=%s -o %s", workingPath, covFile, outFile)
	if _, err := utils.RunShellCmd(c.sh, "-c", cmd); err != nil {
		return fmt.Errorf("GoToolCreateCoverHTMLReport run command failed: %s", cmd)
	}
	return nil
}

// CreateDiffCoverHTMLReport .
func (c *ShCmd) CreateDiffCoverHTMLReport(workingPath, covFile string) error {
	xmlOutput := strings.Replace(covFile, filepath.Ext(covFile), "xml", 1)
	covCmd := fmt.Sprintf("cd %s; gocov convert %s | gocov-xml > %s", workingPath, covFile, xmlOutput)
	if _, err := utils.RunShellCmd(c.sh, "-c", covCmd); err != nil {
		return fmt.Errorf("CreateDiffCoverHTMLReport run command failed: %s", covCmd)
	}

	htmlOutput := fmt.Sprintf("%s_diff.html", getFileNameWithoutExt(covFile))
	diffCoverCmd := fmt.Sprintf("cd %s; diff-cover %s --compare-branch=master --html-report=%s", workingPath, xmlOutput, htmlOutput)
	if _, err := utils.RunShellCmd(c.sh, "-c", diffCoverCmd); err != nil {
		return fmt.Errorf("CreateDiffCoverHTMLReport run command failed: %s", diffCoverCmd)
	}
	return nil
}
