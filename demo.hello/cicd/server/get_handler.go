package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
)

// GetStoreIssues .
func GetStoreIssues(c echo.Context) error {
	key := c.QueryParam("storeKey")
	tree, ok := TreeMap[key]
	if !ok {
		return c.String(http.StatusOK, fmt.Sprintf("StoreKey [%s] not found.\n", key))
	}

	for tree.IsRunning() {
		time.Sleep(time.Duration(500) * time.Millisecond)
	}
	return c.String(http.StatusOK, tree.ToText())
}

// GetSingleIssue .
func GetSingleIssue(c echo.Context) error {
	key := c.QueryParam("storeKey")
	issueKey := c.QueryParam("key")

	tree, ok := TreeMap[key]
	if !ok {
		return c.String(http.StatusOK, fmt.Sprintf("Store [%s] not found.\n", key))
	}

	for tree.IsRunning() {
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	outLines := make([]string, 10)
	issue, text := tree.GetIssueAndText(issueKey, "")
	outLines = append(outLines, text)
	for _, subIssueKey := range issue.SubIssues {
		_, subText := tree.GetIssueAndText(subIssueKey, "\t")
		outLines = append(outLines, subText)
	}
	return c.String(http.StatusOK, strings.Join(outLines, ""))
}
