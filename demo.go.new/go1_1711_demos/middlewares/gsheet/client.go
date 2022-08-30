package gsheet

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

//
// Gsheets
//

const (
	scopeReadOnly  = "https://www.googleapis.com/auth/spreadsheets.readonly"
	scopeReadWrite = "https://www.googleapis.com/auth/spreadsheets"
)

var (
	gSheets     *GSheets
	gSheetsOnce sync.Once

	ErrGsheetsNoDataFound = fmt.Errorf("No data found")
)

type GsheetsParam struct {
	SpreadSheetId    string
	SheetName        string
	SheetId          int64
	RangeName        string
	StartRowIndex    int64
	EndRowIndex      int64
	StartColumnIndex int64
	EndColumnIndex   int64
}

func (param GsheetsParam) getFullRangeName() string {
	return fmt.Sprintf("%s!%s", param.SheetName, param.RangeName)
}

type GSheets struct {
	srv *sheets.Service
}

func NewGSheets() *GSheets {
	gSheetsOnce.Do(func() {
		b, err := ioutil.ReadFile(filepath.Join(os.Getenv("PRJ_PATH"), "internal", "credentials.json"))
		if err != nil {
			log.Fatalf("Unable to read client secret file: %v", err)
		}

		config, err := google.ConfigFromJSON(b, scopeReadWrite)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client := getClient(config)
		srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Fatalf("Unable to retrieve Sheets client: %v", err)
		}
		gSheets = &GSheets{
			srv: srv,
		}
	})

	return gSheets
}

func (gSheets *GSheets) ReadByRange(ctx context.Context, param GsheetsParam) ([][]interface{}, error) {
	resp, err := gSheets.srv.Spreadsheets.Values.Get(param.SpreadSheetId, param.getFullRangeName()).Do()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve data from sheet: %v", err)
	}
	if len(resp.Values) == 0 {
		return nil, ErrGsheetsNoDataFound
	}

	return resp.Values, nil
}

func (gSheets *GSheets) WriteByRange(ctx context.Context, param GsheetsParam, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         values,
	}
	_, err := gSheets.srv.Spreadsheets.Values.Update(param.SpreadSheetId, param.getFullRangeName(), valueRange).
		Context(ctx).ValueInputOption("RAW").Do()
	return err
}

func (gSheets *GSheets) AppendByRange(ctx context.Context, param GsheetsParam, values [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         values,
	}
	_, err := gSheets.srv.Spreadsheets.Values.Append(param.SpreadSheetId, param.getFullRangeName(), valueRange).
		Context(ctx).ValueInputOption("RAW").Do()
	return err
}

// fix ACCESS_TOKEN_SCOPE_INSUFFICIENT: use scopeReadWrite instead of scopeReadOnly, and re-create token.json
func (gSheets *GSheets) CreateSpreadSheet(ctx context.Context, spreadSheetTitle, sheetTitle string) (string, error) {
	spreadSheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: spreadSheetTitle,
		},
		Sheets: []*sheets.Sheet{
			{
				Properties: &sheets.SheetProperties{
					Title: sheetTitle,
				},
			},
		},
	}
	sheet, err := gSheets.srv.Spreadsheets.Create(spreadSheet).Context(ctx).Do()
	if err != nil {
		return "", err
	}

	return sheet.SpreadsheetUrl, nil
}

func (gSheets *GSheets) CreateSheet(ctx context.Context, spreadSheetId, sheetTitle string) (int64, error) {
	req := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: sheetTitle,
			},
		},
	}
	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests:                     []*sheets.Request{req},
		IncludeSpreadsheetInResponse: true,
	}
	resp, err := gSheets.srv.Spreadsheets.BatchUpdate(spreadSheetId, batchReq).Context(ctx).Do()
	if err != nil {
		return -1, err
	}

	allSheets := resp.UpdatedSpreadsheet.Sheets
	newSheet := allSheets[len(allSheets)-1]
	return newSheet.Properties.SheetId, err
}

func (gSheets *GSheets) GetSpreadSheetInfo(ctx context.Context, spreadSheetId string) (*sheets.Spreadsheet, error) {
	resp, err := gSheets.srv.Spreadsheets.Get(spreadSheetId).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateCellsStyle: refer https://developers.google.com/sheets/api/samples/formatting
func (gSheets *GSheets) UpdateCellsStyle(ctx context.Context, param GsheetsParam) error {
	req := &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          param.SheetId,
				StartRowIndex:    param.StartRowIndex,
				EndRowIndex:      param.EndRowIndex,
				StartColumnIndex: param.StartColumnIndex,
				EndColumnIndex:   param.EndColumnIndex,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					TextFormat: &sheets.TextFormat{
						FontFamily: "Verdana",
						FontSize:   9,
					},
				},
			},
			Fields: "userEnteredFormat(textFormat)",
		},
	}
	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}
	_, err := gSheets.srv.Spreadsheets.BatchUpdate(param.SpreadSheetId, batchReq).Context(ctx).Do()
	return err
}

//
// Gsheet Auth
// https://developers.google.com/sheets/api/quickstart/go
//
// 1. Open Google Cloud Console
// 2. Create a Google Cloud Platform project
// 3. Enable Google Sheets Api
// 4. Create a OAuth Client Id (type:Desktop), and downlaod to credentials.json
//

func getClient(config *oauth2.Config) *http.Client {
	prjPath, ok := os.LookupEnv("PRJ_PATH")
	if !ok {
		log.Fatalln("env [PRJ_PATH] is not set")
	}

	tokFile := filepath.Join(prjPath, "internal", "token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func saveToken(path string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}

	if err := json.NewEncoder(f).Encode(token); err != nil {
		log.Fatalf("Json encode error: %v", err)
	}
}

//
// Helper
//

func prettyPrintSpreadSheetMeta(spreadSheet *sheets.Spreadsheet) {
	fmt.Printf("title=[%s],url=%s\n", spreadSheet.Properties.Title, spreadSheet.SpreadsheetUrl)
	for _, sheet := range spreadSheet.Sheets {
		sProperties := sheet.Properties
		fmt.Printf("\tsheet_id=%d,sheet_title=%s\n", sProperties.SheetId, sProperties.Title)
		fmt.Printf("\trow_count=%d,column_count=%d\n", sProperties.GridProperties.RowCount, sProperties.GridProperties.ColumnCount)
	}
}

func prettyPrintRespRows(rows [][]interface{}) {
	for _, row := range rows {
		fields := make([]string, 0, len(row))
		for _, field := range row {
			fields = append(fields, fmt.Sprintf("%v", field))
		}
		fmt.Println(strings.Join(fields, "||"))
	}
}
