package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"demo.hello/apps/cicd/pkg"
	"github.com/labstack/echo"
)

var (
	// StoreCtx .
	StoreCtx context.Context
	// StoreCancel .
	StoreCancel context.CancelFunc
)

// StoreUsage .
func StoreUsage(c echo.Context) error {
	releaseCycle := c.QueryParam("releaseCycle")
	tree, ok := treeMap[releaseCycle]
	if !ok {
		return c.String(http.StatusOK, fmt.Sprintf("Release Cycle [%s] not found.\n", releaseCycle))
	}
	return c.String(http.StatusOK, tree.UsageToText())
}

// StoreReleaseCycleIssues .
func StoreReleaseCycleIssues(c echo.Context) error {
	releaseCycle := c.QueryParam("releaseCycle")
	jql := fmt.Sprintf(`"Release Cycle" = "%s"`, releaseCycle)

	if _, ok := treeMap[releaseCycle]; ok {
		forceUpdate := c.QueryParam("forceUpdate")
		if strings.ToLower(forceUpdate) == "true" {
			StoreCancel()
			delete(treeMap, releaseCycle)
		} else {
			return c.String(http.StatusOK, fmt.Sprintf("Release Cycle [%s] exist.", releaseCycle))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Search issues [%s] failed: %v\n", releaseCycle, err))
	}

	StoreCtx, StoreCancel = context.WithCancel(context.Background())
	tree := pkg.NewJiraIssuesTree(StoreCtx, 8)
	treeMap[releaseCycle] = tree
	tree.Collect()
	for _, key := range keys {
		tree.SubmitIssue(key)
	}
	return c.String(http.StatusOK, fmt.Sprintf("Issues for [%s] stored.", releaseCycle))
}
