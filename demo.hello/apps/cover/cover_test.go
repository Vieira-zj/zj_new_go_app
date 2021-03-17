package cover

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

/*
Refer: https://github.com/golang/tools.git
*/

var projectPath = filepath.Join(os.Getenv("HOME"), "Workspaces/zj_repos/zj_go2_project")

func TestFindFuncs(t *testing.T) {
	// 通过ast来解析go文件中的func
	path := filepath.Join(projectPath, "demo.hello/apps/cover/test/calculator.go")
	fes, err := findFuncs(path)
	if err != nil {
		t.Fatal(err)
	}

	for _, fe := range fes {
		fmt.Printf("%+v\n", *fe)
	}
}

func TestParseProfiles(t *testing.T) {
	// 解析 .cov 文件, 生成go文件profile结构化数据
	path := filepath.Join(projectPath, "demo.hello/apps/cover/reports/cov_results.txt")
	profiles, err := parseProfiles(path)
	if err != nil {
		t.Fatal(err)
	}

	for _, p := range profiles {
		fmt.Printf("\nfilename=%s,mode=%s,blocks:\n", p.FileName, strings.Trim(p.Mode, " "))
		for _, b := range p.Blocks {
			fmt.Printf("%+v\n", b)
		}
	}
}

func TestFuncCoverOutput(t *testing.T) {
	// 解析 .cov 文件, 输出函数级的覆盖率数据
	path := filepath.Join(projectPath, "demo.hello/apps/cover/reports/cov_results.txt")
	if err := funcCoverOutput(path); err != nil {
		t.Fatal(err)
	}
}
