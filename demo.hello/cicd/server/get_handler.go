package server

import (
	"fmt"
	"net/http"
	"strings"

	"demo.hello/cicd/pkg"
	"github.com/labstack/echo"
)

// GetStoreIssues .
func GetStoreIssues(c echo.Context) error {
	req, err := parseBodyToIssuesHandlerReq(c.Request().Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	locker.RLock()
	tree, ok := TreeMap[req.StoreKey]
	locker.RUnlock()
	if !ok {
		fmt.Printf("Store [%s] not found and try to create.\n", req.StoreKey)
		jql, err := getJQLFromReq(req)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		tree, err = storeJQLIssues(req.StoreKey, jql, req.ForceUpdate)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	tree.WaitDone()
	return c.String(http.StatusOK, pkg.GetIssuesTreeText(tree))
}

// GetSingleIssue .
func GetSingleIssue(c echo.Context) error {
	req, err := parseBodyToIssuesHandlerReq(c.Request().Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	locker.RLock()
	tree, ok := TreeMap[req.StoreKey]
	locker.RUnlock()
	if !ok {
		return c.String(http.StatusOK, fmt.Sprintf("Store [%s] not found.\n", req.StoreKey))
	}
	tree.WaitDone()

	outLines := make([]string, 10)
	issue, text := pkg.GetIssueAndMRsText(tree, req.IssueKey, "")
	outLines = append(outLines, text)
	for _, subIssueKey := range issue.SubIssues {
		_, subText := pkg.GetIssueAndMRsText(tree, subIssueKey, "\t")
		outLines = append(outLines, subText)
	}
	return c.String(http.StatusOK, strings.Join(outLines, ""))
}

// GetRepos .
func GetRepos(c echo.Context) error {
	tree, err := getStore(c)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	content := pkg.GetDeployReposText(tree) + "\n" + pkg.GetShortDeployReposText(tree)
	return c.String(http.StatusOK, content)
}

// StoreUsage .
func StoreUsage(c echo.Context) error {
	tree, err := getStore(c)
	if err != nil {
		return c.String(http.StatusOK, err.Error())
	}
	content := pkg.GetIssuesTreeUsageText(tree) + "\n" + pkg.GetIssuesTreeSummaryText(tree)
	return c.String(http.StatusOK, content)
}

func getStore(c echo.Context) (pkg.Tree, error) {
	locker.RLock()
	defer locker.RUnlock()

	req, err := parseBodyToIssuesHandlerReq(c.Request().Body)
	if err != nil {
		return nil, err
	}

	tree, ok := TreeMap[req.StoreKey]
	if !ok {
		return nil, fmt.Errorf("Store [%s] not found", req.StoreKey)
	}
	tree.WaitDone()
	return tree, nil
}
