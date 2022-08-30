package gsheet

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	prjPath := filepath.Join(os.Getenv("HOME"), "Workspaces/zj_repos/zj_new_go_project/demo.go.new/go1_1711_demos")
	os.Setenv("PRJ_PATH", prjPath)
	m.Run()
}

// https://docs.google.com/spreadsheets/d/{SpreadSheet_Id}/edit#gid={Sheet_Id}
var sheetIdForTest = ""

func TestGSheetsCreateSpreadSheet(t *testing.T) {
	t.Skip("Run once")
	gsheets := NewGSheets()
	spreadSheetTitle := "Test: create gsheet api"
	spreadSheetUrl, err := gsheets.CreateSpreadSheet(context.Background(), spreadSheetTitle, "test-01")
	assert.NoError(t, err)
	t.Log("spreadsheet created:", spreadSheetUrl)
}

func TestGSheetsCreateSheet(t *testing.T) {
	gsheets := NewGSheets()
	sheetId, err := gsheets.CreateSheet(context.Background(), sheetIdForTest, "test-11")
	assert.NoError(t, err)
	t.Log("sheet created:", sheetId)
}

func TestGsheetsReadByRange(t *testing.T) {
	gsheets := NewGSheets()
	gSheetParam := GsheetsParam{
		SpreadSheetId: sheetIdForTest,
		SheetName:     "test-01",
		RangeName:     "A2:D4",
	}
	rows, err := gsheets.ReadByRange(context.Background(), gSheetParam)
	assert.NoError(t, err)
	prettyPrintRespRows(rows)
	t.Log("read finish")
}

func TestGSheetsWriteByRange(t *testing.T) {
	gsheets := NewGSheets()
	assert.NotNil(t, gsheets)

	gSheetParam := GsheetsParam{
		SpreadSheetId: sheetIdForTest,
		SheetName:     "test-11",
		RangeName:     "A1:D6",
	}
	values := [][]interface{}{
		{"AppName", "Path", "Request", "Count"},
		{"app-test01", "/index", `{"msg":"hello"}`, 10},
		{"app-test02", "/ping", `{"msg":"pong"}`, 3},
	}
	err := gsheets.WriteByRange(context.Background(), gSheetParam, values)
	assert.NoError(t, err)
	t.Log("write finish")
}

func TestGSheetsAppendByRange(t *testing.T) {
	gsheets := NewGSheets()
	assert.NotNil(t, gsheets)

	gSheetParam := GsheetsParam{
		SpreadSheetId: sheetIdForTest,
		SheetName:     "test-11",
		RangeName:     "A1:D6",
	}
	values := [][]interface{}{
		{"app-test03", "/index", `{"msg":"hello"}`, 11},
		{"app-test04", "/ping", `{"msg":"pong"}`, 7},
	}
	err := gsheets.AppendByRange(context.Background(), gSheetParam, values)
	assert.NoError(t, err)
	t.Log("append finish")
}

func TestGetSpreadSheetInfo(t *testing.T) {
	gsheets := NewGSheets()
	spreadSheet, err := gsheets.GetSpreadSheetInfo(context.Background(), sheetIdForTest)
	assert.NoError(t, err)
	prettyPrintSpreadSheetMeta(spreadSheet)
}

func TestGSheetsUpdateCellsStyle(t *testing.T) {
	gsheets := NewGSheets()
	gSheetParam := GsheetsParam{
		SpreadSheetId:    sheetIdForTest,
		SheetId:          530966751,
		StartRowIndex:    0,
		EndRowIndex:      10,
		StartColumnIndex: 0,
		EndColumnIndex:   4,
	}
	err := gsheets.UpdateCellsStyle(context.Background(), gSheetParam)
	assert.NoError(t, err)
	t.Log("update cell style finish")
}
