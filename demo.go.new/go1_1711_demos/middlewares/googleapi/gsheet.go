package googleapi

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

//
// Gsheets
//

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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client := getAuthClient(sheets.SpreadsheetsScope)
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
	resp, err := gSheets.srv.Spreadsheets.Values.Get(param.SpreadSheetId, param.getFullRangeName()).Context(ctx).Do()
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
func (gSheets *GSheets) CreateSpreadSheet(ctx context.Context, spreadSheetTitle, sheetTitle string) (string, int64, error) {
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
	resp, err := gSheets.srv.Spreadsheets.Create(spreadSheet).Context(ctx).Do()
	if err != nil {
		return "", -1, err
	}

	return resp.SpreadsheetId, resp.Sheets[0].Properties.SheetId, nil
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
		Requests: []*sheets.Request{req},
		// use resp.UpdatedSpreadsheet
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

func (gSheets *GSheets) GetSpreadSheet(ctx context.Context, spreadSheetId string) (*sheets.Spreadsheet, error) {
	resp, err := gSheets.srv.Spreadsheets.Get(spreadSheetId).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateCellsStyle: refer https://developers.google.com/sheets/api/samples/formatting
func (gSheets *GSheets) AddStyleFoSheet(ctx context.Context, param GsheetsParam) error {
	req1 := &sheets.Request{
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

	req2 := &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          param.SheetId,
				StartRowIndex:    0,
				EndRowIndex:      1,
				StartColumnIndex: 0,
				EndColumnIndex:   4,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					TextFormat: &sheets.TextFormat{
						Bold: true,
					},
					BackgroundColor: &sheets.Color{
						Red:   0.73,
						Green: 0.35,
						Blue:  0.41,
						Alpha: 0.5,
					},
				},
			},
			Fields: "userEnteredFormat(textFormat,backgroundColor)",
		},
	}

	req3 := &sheets.Request{
		UpdateBorders: &sheets.UpdateBordersRequest{
			Range: &sheets.GridRange{
				SheetId:          param.SheetId,
				StartRowIndex:    param.StartRowIndex,
				EndRowIndex:      param.EndRowIndex,
				StartColumnIndex: param.StartColumnIndex,
				EndColumnIndex:   param.EndColumnIndex,
			},
			Top: &sheets.Border{
				Style: "SOLID",
			},
			Bottom: &sheets.Border{
				Style: "SOLID",
			},
			Left: &sheets.Border{
				Style: "SOLID",
			},
			Right: &sheets.Border{
				Style: "SOLID",
			},
			InnerHorizontal: &sheets.Border{
				Style: "SOLID",
			},
			InnerVertical: &sheets.Border{
				Style: "SOLID",
			},
		},
	}

	req4 := &sheets.Request{
		UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
			Properties: &sheets.SheetProperties{
				SheetId: param.SheetId,
				GridProperties: &sheets.GridProperties{
					FrozenRowCount: 1,
				},
			},
			Fields: "gridProperties.frozenRowCount",
		},
	}

	batchReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req1, req2, req3, req4},
	}
	_, err := gSheets.srv.Spreadsheets.BatchUpdate(param.SpreadSheetId, batchReq).Context(ctx).Do()
	return err
}

//
// Helper
//

// GetSheetUrl: format as https://docs.google.com/spreadsheets/d/{SpreadSheet_Id}/edit#gid={Sheet_Id}
func GetSheetUrl(spreadSheetId string, sheetId int64) string {
	return fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit#gid=%d", spreadSheetId, sheetId)
}

func PrettyPrintSpreadSheetMeta(spreadSheet *sheets.Spreadsheet) {
	fmt.Printf("title=[%s],url=%s\n", spreadSheet.Properties.Title, spreadSheet.SpreadsheetUrl)
	for _, sheet := range spreadSheet.Sheets {
		sProperties := sheet.Properties
		fmt.Printf("\tsheet_id=%d,sheet_title=%s\n", sProperties.SheetId, sProperties.Title)
		fmt.Printf("\trow_count=%d,column_count=%d\n", sProperties.GridProperties.RowCount, sProperties.GridProperties.ColumnCount)
	}
}

func PrettyPrintRespRows(rows [][]interface{}) {
	for _, row := range rows {
		fields := make([]string, 0, len(row))
		for _, field := range row {
			fields = append(fields, fmt.Sprintf("%v", field))
		}
		fmt.Println(strings.Join(fields, "|-|"))
	}
}
