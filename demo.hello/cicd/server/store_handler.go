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
	// StoreCancelMap .
	StoreCancelMap map[string]context.CancelFunc = make(map[string]context.CancelFunc)
)

// StoreReleaseCycleIssues .
func StoreReleaseCycleIssues(c echo.Context) error {
	releaseCycle := c.QueryParam("releaseCycle")
	forceUpdate := c.QueryParam("forceUpdate")
	jql := fmt.Sprintf(`"Release Cycle" = "%s"`, releaseCycle)

	retContent := storeJQLIssues(releaseCycle, jql, forceUpdate)
	return c.String(http.StatusOK, retContent)
}

func storeJQLIssues(key, jql, forceUpdate string) string {
	if _, ok := treeMap[key]; ok {
		if strings.ToLower(forceUpdate) == "true" {
			cancel := StoreCancelMap[key]
			cancel()
			delete(treeMap, key)
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
	treeMap[key] = tree
	tree.Collect()
	for _, key := range keys {
		tree.SubmitIssue(key)
	}
	return fmt.Sprintf("Issues for key [%s] stored.", key)
}

// StoreUsage .
func StoreUsage(c echo.Context) error {
	key := c.QueryParam("storeKey")
	tree, ok := treeMap[key]
	if !ok {
		return c.String(http.StatusOK, fmt.Sprintf("StoreKey [%s] not found.\n", key))
	}
	return c.String(http.StatusOK, tree.UsageToText())
}
