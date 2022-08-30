package main

import (
	"context"
	"fmt"
	"go1_1711_demo/middlewares/gsheet"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	// gSheetTest()

	ver := runtime.Version()
	fmt.Printf("%s demo\n", ver)
}

func gSheetTest() {
	prjPath := filepath.Join(os.Getenv("HOME"), "Workspaces/zj_repos/zj_new_go_project/demo.go.new/go1_1711_demos")
	os.Setenv("PRJ_PATH", prjPath)

	gsheets := gsheet.NewGSheets()
	title := "Test: create gsheet api"
	spreadSheetId, err := gsheets.CreateSpreadSheet(context.Background(), title, "test-01")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("spreadsheet created:", spreadSheetId)
}
