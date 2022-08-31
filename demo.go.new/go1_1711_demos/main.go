package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"go1_1711_demo/middlewares/googleapi"
)

func main() {
	// gDriverTest()

	ver := runtime.Version()
	fmt.Printf("%s demo\n", ver)
}

func gDriverTest() {
	prjPath := filepath.Join(os.Getenv("HOME"), "Workspaces/zj_repos/zj_new_go_project/demo.go.new/go1_1711_demos")
	os.Setenv("PRJ_PATH", prjPath)

	gdriver := googleapi.NewGDriver()
	gdriver.ListFilesSample(context.Background())
}
