package pkg

import (
	"fmt"
	"log"
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

const (
	coverRptTypeFunc = "func"
	coverRptTypeHTML = "html"
)

// GoToolCreateCoverFuncReport .
func (c *ShCmd) GoToolCreateCoverFuncReport(workingPath, covFile string) (string, error) {
	return c.goToolCreateCoverReport(workingPath, covFile, coverRptTypeFunc)
}

// GoToolCreateCoverHTMLReport .
func (c *ShCmd) GoToolCreateCoverHTMLReport(workingPath, covFile string) (string, error) {
	return c.goToolCreateCoverReport(workingPath, covFile, coverRptTypeHTML)
}

func (c *ShCmd) goToolCreateCoverReport(workingPath, covFile, coverType string) (string, error) {
	outFile := strings.Replace(covFile, filepath.Ext(covFile), "."+coverType, 1)
	cmd := fmt.Sprintf("cd %s; go tool cover -%s=%s -o %s", workingPath, coverType, covFile, outFile)
	log.Println("Run cmd:", cmd)
	output, err := utils.RunShellCmd(c.sh, "-c", cmd)
	if err != nil {
		return "", fmt.Errorf("goToolCreateCoverReport run command error: %s", err)
	}
	return output, nil
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
