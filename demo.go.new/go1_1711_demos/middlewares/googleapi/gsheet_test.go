package googleapi

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

var sheetIdForTest = ""

func TestGSheetsCreateSpreadSheet(t *testing.T) {
	t.Skip("Run once")
	gsheets := NewGSheets()
	spreadSheetTitle := "Test: create gsheet api"
	spreadSheetId, sheetId, err := gsheets.CreateSpreadSheet(context.Background(), spreadSheetTitle, "test-01")
	assert.NoError(t, err)
	t.Log("spreadsheet created:", GetSheetUrl(spreadSheetId, sheetId))
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
	PrettyPrintRespRows(rows)
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

func TestGetSpreadSheet(t *testing.T) {
	gsheets := NewGSheets()
	spreadSheet, err := gsheets.GetSpreadSheet(context.Background(), sheetIdForTest)
	assert.NoError(t, err)
	PrettyPrintSpreadSheetMeta(spreadSheet)
}

func TestAddStyleFoSheet(t *testing.T) {
	gsheets := NewGSheets()
	gSheetParam := GsheetsParam{
		SpreadSheetId:    sheetIdForTest,
		SheetId:          1419638410,
		StartRowIndex:    0,
		EndRowIndex:      10,
		StartColumnIndex: 0,
		EndColumnIndex:   4,
	}
	err := gsheets.AddStyleFoSheet(context.Background(), gSheetParam)
	assert.NoError(t, err)
	t.Log("add style finish")
}
