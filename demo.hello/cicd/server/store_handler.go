package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

// StoreIssuesReq .
type StoreIssuesReq struct {
	ReleaseCycle string `json:"releaseCycle"`
	FixVersion   string `json:"fixVersion"`
	Query        string `json:"query"`
	ForceUpdate  bool   `json:"forceUpdate"`
}

// StoreIssues .
func StoreIssues(c echo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Read request body failed.")
	}

	req := &StoreIssuesReq{}
	if err := json.Unmarshal(body, &req); err != nil {
		return c.String(http.StatusInternalServerError, "Unmarshal body failed.")
	}

	if len(req.ReleaseCycle) > 0 {
		jql := fmt.Sprintf(`"Release Cycle" = "%s"`, req.ReleaseCycle)
		return c.String(http.StatusOK, storeJQLIssues(req.ReleaseCycle, jql, req.ForceUpdate))
	}
	if len(req.FixVersion) > 0 {
		jql := fmt.Sprintf("fixVersion = %s", req.FixVersion)
		return c.String(http.StatusOK, storeJQLIssues(req.FixVersion, jql, req.ForceUpdate))
	}
	if len(req.Query) > 0 {
		return c.String(http.StatusOK, storeJQLIssues(req.Query, req.Query, req.ForceUpdate))
	}
	return c.String(http.StatusBadRequest, fmt.Sprintln("No query found."))
}

func storeJQLIssues(key, jql string, forceUpdate bool) string {
	if _, ok := TreeMap[key]; ok {
		if forceUpdate {
			fmt.Printf("Force update for store [%s].\n", key)
			StoreCancelMap[key]()
			delete(TreeMap, key)
		} else {
			return fmt.Sprintf("Store [%s] already exist.\n", key)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	issues, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		return fmt.Sprintf("Search issues by jql [%s] failed: %v\n", jql, err)
	}

	ctx, cancel = context.WithCancel(context.Background())
	StoreCancelMap[key] = cancel
	tree := pkg.NewJiraIssuesTree(ctx, 8)
	TreeMap[key] = tree
	for _, issueID := range issues {
		tree.SubmitIssue(issueID)
	}
	return fmt.Sprintf("Store [%s] saved.\n", key)
}
