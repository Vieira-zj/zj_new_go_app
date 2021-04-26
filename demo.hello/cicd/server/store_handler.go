package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"demo.hello/cicd/pkg"
	"github.com/labstack/echo"
)

// IssuesHandlerReq .
type IssuesHandlerReq struct {
	StoreKey     string `json:"storeKey"`
	IssueKey     string `json:"issueKey"`
	StoreKeyType string `json:"storeKeyType"`
	ForceUpdate  bool   `json:"forceUpdate"`
}

// StoreIssues .
func StoreIssues(c echo.Context) error {
	req, err := parseBodyToIssuesHandlerReq(c.Request().Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	jql, err := getJQLFromReq(req)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	_, err = storeJQLIssues(req.StoreKey, jql, req.ForceUpdate)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, fmt.Sprintf("Store [%s] saved.\n", req.StoreKey))
}

func parseBodyToIssuesHandlerReq(reader io.ReadCloser) (*IssuesHandlerReq, error) {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.New("Read request body failed")
	}
	req := &IssuesHandlerReq{}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, errors.New("Unmarshal body failed")
	}
	return req, nil
}

func getJQLFromReq(req *IssuesHandlerReq) (string, error) {
	var jql string
	if req.StoreKeyType == typeReleaseCycle {
		jql = fmt.Sprintf(`"Release Cycle" = "%s"`, req.StoreKey)
	} else if req.StoreKeyType == typeFixVersion {
		jql = fmt.Sprintf("fixVersion = %s", req.StoreKey)
	} else if req.StoreKeyType == typeJQL {
		jql = req.StoreKey
	} else {
		return "", fmt.Errorf("Invalid store key type: %s", req.StoreKeyType)
	}

	if len(jql) == 0 {
		return "", errors.New("No query found in request")
	}
	return jql, nil
}

func storeJQLIssues(key, jql string, forceUpdate bool) (pkg.Tree, error) {
	locker.Lock()
	defer locker.Unlock()

	if _, ok := TreeMap[key]; ok {
		if forceUpdate {
			fmt.Printf("Force update for store [%s].\n", key)
			delete(TreeMap, key)
		} else {
			return nil, fmt.Errorf("Store [%s] already exist", key)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(searchTimeout)*time.Second)
	defer cancel()
	issues, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		return nil, fmt.Errorf("Search issues by jql [%s] failed: %v", jql, err)
	}

	tree := pkg.NewJiraIssuesTreeV2(parallel)
	TreeMap[key] = tree
	for _, issueID := range issues {
		tree.SubmitIssue(issueID)
	}
	return tree, nil
}
