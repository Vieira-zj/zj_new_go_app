package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"demo.hello/cicd/pkg"
	"github.com/labstack/echo"
)

// GetStoreIssuesReq .
type GetStoreIssuesReq struct {
	StoreKey string `json:"storeKey"`
	IssueKey string `json:"issueKey"`
}

// GetStoreIssues .
func GetStoreIssues(c echo.Context) error {
	req, err := parseBodyToGetStoreIssuesReq(c.Request().Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	tree, ok := TreeMap[req.StoreKey]
	if !ok {
		return c.String(http.StatusOK, fmt.Sprintf("StoreKey [%s] not found.\n", req.StoreKey))
	}

	tree.WaitDone()
	return c.String(http.StatusOK, pkg.GetIssuesTreeText(tree))
}

// GetSingleIssue .
func GetSingleIssue(c echo.Context) error {
	req, err := parseBodyToGetStoreIssuesReq(c.Request().Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	tree, ok := TreeMap[req.StoreKey]
	if !ok {
		return c.String(http.StatusOK, fmt.Sprintf("Store [%s] not found.\n", req.StoreKey))
	}

	tree.WaitDone()
	outLines := make([]string, 10)
	issue, text := pkg.GetIssueAndText(tree, req.IssueKey, "")
	outLines = append(outLines, text)
	for _, subIssueKey := range issue.SubIssues {
		_, subText := pkg.GetIssueAndText(tree, subIssueKey, "\t")
		outLines = append(outLines, subText)
	}
	return c.String(http.StatusOK, strings.Join(outLines, ""))
}

// StoreUsage .
func StoreUsage(c echo.Context) error {
	req, err := parseBodyToGetStoreIssuesReq(c.Request().Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	tree, ok := TreeMap[req.StoreKey]
	if !ok {
		return c.String(http.StatusOK, fmt.Sprintf("StoreKey [%s] not found.\n", req.StoreKey))
	}
	return c.String(http.StatusOK, pkg.GetIssuesTreeUsageText(tree))
}

func parseBodyToGetStoreIssuesReq(reader io.ReadCloser) (*GetStoreIssuesReq, error) {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Read request body failed")
	}

	req := &GetStoreIssuesReq{}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, errors.New("Unmarshal body failed")
	}
	return req, nil
}
