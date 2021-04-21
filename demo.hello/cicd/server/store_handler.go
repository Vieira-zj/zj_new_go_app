package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"demo.hello/cicd/pkg"
	"github.com/labstack/echo"
)

var (
	// TreeMap .
	TreeMap = make(map[string]*pkg.JiraIssuesTree)
	// StoreCancelMap .
	StoreCancelMap map[string]context.CancelFunc = make(map[string]context.CancelFunc)
)

// StoreIssues .
func StoreIssues(c echo.Context) error {
	var key, jql string
	releaseCycle := c.QueryParam("releaseCycle")
	fixVersion := c.QueryParam("fixVersion")
	query := c.QueryParam("query")
	forceUpdate := c.QueryParam("forceUpdate")

	if len(releaseCycle) > 0 {
		key = releaseCycle
		jql = fmt.Sprintf(`"Release Cycle" = "%s"`, releaseCycle)
	} else if len(fixVersion) > 0 {
		key = fixVersion
		jql = fmt.Sprintf("fixVersion = %s", fixVersion)
	} else if len(query) > 0 {
		key = query
		jql = query
	} else {
		return c.String(http.StatusBadRequest, fmt.Sprintln("no query found."))
	}

	retContent := storeJQLIssues(key, jql, forceUpdate)
	return c.String(http.StatusOK, retContent)
}

func storeJQLIssues(key, jql, forceUpdate string) string {
	if _, ok := TreeMap[key]; ok {
		if strings.ToLower(forceUpdate) == "true" {
			cancel := StoreCancelMap[key]
			cancel()
			delete(TreeMap, key)
		} else {
			return fmt.Sprintf("Store for key [%s] exist.", key)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		return fmt.Sprintf("Search issues for jql [%s] failed: %v\n", jql, err)
	}

	ctx, cancel = context.WithCancel(context.Background())
	StoreCancelMap[key] = cancel
	tree := pkg.NewJiraIssuesTree(ctx, 8)
	TreeMap[key] = tree
	tree.Collect()
	for _, key := range keys {
		tree.SubmitIssue(key)
	}
	return fmt.Sprintf("Issues for key [%s] stored.", key)
}

// StoreUsage .
func StoreUsage(c echo.Context) error {
	key := c.QueryParam("storeKey")
	tree, ok := TreeMap[key]
	if !ok {
		return c.String(http.StatusOK, fmt.Sprintf("StoreKey [%s] not found.\n", key))
	}
	return c.String(http.StatusOK, tree.UsageToText())
}
