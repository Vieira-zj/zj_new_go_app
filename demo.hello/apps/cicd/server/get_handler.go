package server

import (
	"fmt"
	"net/http"

	"demo.hello/apps/cicd/pkg"
	"github.com/labstack/echo"
)

// GetReleaseCycleIssues .
func GetReleaseCycleIssues(c echo.Context) error {
	releaseCycle := c.QueryParam("releaseCycle")
	tree, ok := treeMap[releaseCycle]
	if !ok {
		return c.String(http.StatusOK, fmt.Sprintf("Release Cycle [%s] not found.\n", releaseCycle))
	}
	return c.String(http.StatusOK, tree.ToText())
}

// GetSingleIssue .
func GetSingleIssue(c echo.Context) error {
	releaseCycle := c.QueryParam("releaseCycle")
	issueKey := c.QueryParam("key")

	tree, ok := treeMap[releaseCycle]
	if !ok {
		return c.String(http.StatusOK, fmt.Sprintf("Release Cycle [%s] not found.\n", releaseCycle))
	}

	value, err := tree.GetStore().Get(issueKey)
	if err != nil {
		return c.String(http.StatusOK, fmt.Sprintf("Get issue [%s] error: %v\n", issueKey, err))
	}
	issue := value.(*pkg.JiraIssue)
	return c.String(http.StatusOK, tree.IssueToText(issue, ""))
}
