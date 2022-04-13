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
func (c *ShCmd) GoToolCreateCoverFuncReport(workingPath, covFilePath string) (string, error) {
	return c.goToolCreateCoverReport(workingPath, covFilePath, coverRptTypeFunc)
}

// GoToolCreateCoverHTMLReport .
func (c *ShCmd) GoToolCreateCoverHTMLReport(workingPath, covFile string) (string, error) {
	return c.goToolCreateCoverReport(workingPath, covFile, coverRptTypeHTML)
}

func (c *ShCmd) goToolCreateCoverReport(workingPath, covFilePath, coverType string) (string, error) {
	outFilePath := GetFilePathWithNewExt(covFilePath, coverType)
	cmd := fmt.Sprintf("cd %s; go tool cover -%s=%s -o %s", workingPath, coverType, covFilePath, outFilePath)
	log.Println("Run cmd:", cmd)
	output, err := utils.RunShellCmd(c.sh, "-c", cmd)
	if err != nil {
		return "", fmt.Errorf("goToolCreateCoverReport run command error: %s", err)
	}

	if coverType == "func" {
		output, err = GetCoverTotalFromFuncReport(outFilePath)
		if err != nil {
			return "", fmt.Errorf("goToolCreateCoverReport error: %w", err)
		}
	}
	return output, nil
}

// GetCoverTotalFromFuncReport .
func GetCoverTotalFromFuncReport(filePath string) (string, error) {
	lines, err := utils.ReadLinesFile(filePath)
	if err != nil {
		return "", fmt.Errorf("GetCoverTotalFromFuncReport read file lines error: %w", err)
	}
	summary := lines[len(lines)-1]
	return getCoverTotalFromSummary(summary), nil
}

// CreateDiffCoverHTMLReport .
func (c *ShCmd) CreateDiffCoverHTMLReport(workingPath, covFilePath string) error {
	xmlOutput := strings.Replace(covFilePath, filepath.Ext(covFilePath), "xml", 1)
	covCmd := fmt.Sprintf("cd %s; gocov convert %s | gocov-xml > %s", workingPath, covFilePath, xmlOutput)
	if _, err := utils.RunShellCmd(c.sh, "-c", covCmd); err != nil {
		return fmt.Errorf("CreateDiffCoverHTMLReport run command failed: %s", covCmd)
	}

	htmlOutput := fmt.Sprintf("%s_diff.html", getFilePathWithoutExt(covFilePath))
	diffCoverCmd := fmt.Sprintf("cd %s; diff-cover %s --compare-branch=master --html-report=%s", workingPath, xmlOutput, htmlOutput)
	if _, err := utils.RunShellCmd(c.sh, "-c", diffCoverCmd); err != nil {
		return fmt.Errorf("CreateDiffCoverHTMLReport run command failed: %s", diffCoverCmd)
	}
	return nil
}
