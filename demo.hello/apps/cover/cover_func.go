package cover

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
)

// funcCoverOutput prints profile data of func level.
func funcCoverOutput(covFilePath string) error {
	// .cov -> profiles -> profile -> .go file, blocks
	profiles, err := parseProfiles(covFilePath)
	if err != nil {
		return err
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	tabber := tabwriter.NewWriter(out, 1, 8, 1, '\t', 0)
	defer tabber.Flush()

	var total, covered int64
	projectPath := filepath.Join(os.Getenv("HOME"), "Workspaces/zj_repos/zj_go2_project")
	// profile -> .go file, blocks
	for _, profile := range profiles {
		fileName := profile.FileName
		filePath := filepath.Join(projectPath, fileName)
		// .go file -> ast -> []FuncExtent
		fes, err := findFuncs(filePath)
		if err != nil {
			return err
		}

		// fe -> multiple profile blocks
		for _, fe := range fes {
			c, t := fe.getCoverage(profile)
			fmt.Fprintf(tabber, "%s:%d:\t%s\t%.1f%%\n", fileName, fe.startLine, fe.name, 100.0*float64(c)/float64(t))
			total += t
			covered += c
		}
	}
	fmt.Fprintf(tabber, "total:\t(statements)\t%.1f%%\n", 100.0*float64(covered)/float64(total))

	return nil
}
