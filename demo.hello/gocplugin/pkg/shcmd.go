package pkg

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"demo.hello/utils"
)

const (
	goBin  = "go"
	gocBin = "goc"
)

var (
	shCmd     *ShCmd
	shCmdOnce sync.Once
)

// ShCmd .
type ShCmd struct {
	sh string
}

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
func (c *ShCmd) GoToolCreateCoverFuncReport(workingPath, covFilePath string) (string, error) {
	return c.goToolCreateCoverReport(workingPath, "", covFilePath, CoverRptTypeFunc)
}

// GoToolCreateCoverHTMLReport .
func (c *ShCmd) GoToolCreateCoverHTMLReport(workingPath, moduleName, covFile string) (string, error) {
	return c.goToolCreateCoverReport(workingPath, moduleName, covFile, CoverRptTypeHTML)
}

func (c *ShCmd) goToolCreateCoverReport(workingPath, moduleName, covFilePath, coverType string) (string, error) {
	outFilePath := GetFilePathWithNewExt(covFilePath, coverType)
	if coverType == CoverRptTypeHTML {
		outFileName := filepath.Base(outFilePath)
		outDirPath := filepath.Join(AppConfig.PublicDir, moduleName)
		if err := utils.MakeDir(outDirPath); err != nil && !errors.Is(err, os.ErrExist) {
			return "", fmt.Errorf("goToolCreateCoverReport make public dir error: %s", outDirPath)
		}
		outFilePath = filepath.Join(outDirPath, outFileName)
	}

	cmd := fmt.Sprintf("cd %s; %s tool cover -%s=%s -o %s", workingPath, goBin, coverType, covFilePath, outFilePath)
	log.Println("Run cmd:", cmd)
	output, err := utils.RunShellCmd(c.sh, "-c", cmd)
	if err != nil {
		return "", fmt.Errorf("goToolCreateCoverReport run command error: %w", err)
	}

	if coverType == CoverRptTypeFunc {
		output, err = getCoverTotalFromFuncReport(outFilePath)
		if err != nil {
			return "", fmt.Errorf("goToolCreateCoverReport error: %w", err)
		}
	}
	return output, nil
}

func getModuleNameFromFileName(fileName string) string {
	name := strings.Split(fileName, ".")[0]
	nameItems := strings.Split(name, "_")
	srvName := strings.Join(nameItems[:len(nameItems)-2], "_")
	meta := GetSrvMetaFromName(srvName)
	return meta.AppName
}

func getCoverTotalFromFuncReport(filePath string) (string, error) {
	lines, err := utils.ReadLinesFile(filePath)
	if err != nil {
		return "", fmt.Errorf("GetCoverTotalFromFuncReport read lines from file error: %w", err)
	}
	summary := lines[len(lines)-1]
	return getCoverTotalFromSummary(summary), nil
}

// GocToolMergeSrvCovers .
func (c *ShCmd) GocToolMergeSrvCovers(covFilePaths []string, mergeFilePath string) error {
	files := strings.Join(covFilePaths, " ")
	mergeCmd := fmt.Sprintf("%s merge %s -o %s", gocBin, files, mergeFilePath)
	if _, err := utils.RunShellCmd(c.sh, "-c", mergeCmd); err != nil {
		return fmt.Errorf("gocToolMergeSrvCovers run command error: %w", err)
	}
	return nil
}

// CreateDiffCoverHTMLReport .
func (c *ShCmd) CreateDiffCoverHTMLReport(workingPath, covFilePath string) error {
	xmlOutput := strings.Replace(covFilePath, filepath.Ext(covFilePath), "xml", 1)
	covCmd := fmt.Sprintf("cd %s; gocov convert %s | gocov-xml > %s", workingPath, covFilePath, xmlOutput)
	if _, err := utils.RunShellCmd(c.sh, "-c", covCmd); err != nil {
		return fmt.Errorf("CreateDiffCoverHTMLReport run command error: %s", covCmd)
	}

	htmlOutput := fmt.Sprintf("%s_diff.html", getFilePathWithoutExt(covFilePath))
	diffCoverCmd := fmt.Sprintf("cd %s; diff-cover %s --compare-branch=master --html-report=%s", workingPath, xmlOutput, htmlOutput)
	if _, err := utils.RunShellCmd(c.sh, "-c", diffCoverCmd); err != nil {
		return fmt.Errorf("CreateDiffCoverHTMLReport run command error: %s", diffCoverCmd)
	}
	return nil
}
