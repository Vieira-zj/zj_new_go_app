package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"demo.hello/utils"
)

/*
jira api docs:
https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/
*/

// JiraTool contains jira rest APIs.
type JiraTool struct {
	restURL  string
	userName string
	userPwd  string
	http     *utils.HTTPUtils
}

// NewJiraTool creates an instance of jira tool.
func NewJiraTool() *JiraTool {
	return &JiraTool{
		restURL:  os.Getenv("JIRA_REST_URL"),
		userName: os.Getenv("JIRA_USER_NAME"),
		userPwd:  os.Getenv("JIRA_USER_PASSWORD"),
		http:     utils.NewHTTPUtils(true),
	}
}

// GetIssue returns an issue by id.
func (jira *JiraTool) GetIssue(ctx context.Context, issueID string, fields []string) ([]byte, error) {
	path := "issue/" + issueID
	if fields != nil && len(fields) > 0 {
		path = path + "?fields=" + strings.Join(fields, ",")
	}
	return jira.get(ctx, path)
}

// Search returns issues by jql.
func (jira *JiraTool) Search(ctx context.Context, jql string, fields []string) ([]byte, error) {
	data := map[string]interface{}{
		"jql":        jql,
		"maxResults": 200,
		"fields":     fields,
	}
	reqData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return jira.post(ctx, "search", string(reqData))
}

// SearchIssues returns issue keys from jql results.
func (jira *JiraTool) SearchIssues(ctx context.Context, jql string) ([]string, error) {
	resp, err := jira.Search(ctx, jql, []string{"key"})
	if err != nil {
		return nil, err
	}

	respMap := make(map[string]interface{})
	err = json.Unmarshal(resp, &respMap)
	if err != nil {
		return nil, err
	}

	total := respMap["total"].(float64)
	fmt.Printf("Search results count: %.0f\n", total)

	issueSlice := respMap["issues"].([]interface{})
	keys := make([]string, 0, len(issueSlice))
	for _, item := range issueSlice {
		issue := item.(map[string]interface{})
		keys = append(keys, issue["key"].(string))
	}
	return keys, nil
}

// GetIssueLink returns issue related links.
func (jira *JiraTool) GetIssueLink(ctx context.Context, issueID string) ([]byte, error) {
	return jira.GetIssue(ctx, issueID, []string{"issuelinks"})
}

// GetRemoteLink returns issue related remote links.
func (jira *JiraTool) GetRemoteLink(ctx context.Context, issueID string) ([]byte, error) {
	path := fmt.Sprintf("issue/%s/remotelink", issueID)
	return jira.get(ctx, path)
}

func (jira *JiraTool) formatPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

func (jira *JiraTool) get(ctx context.Context, path string) ([]byte, error) {
	url := jira.restURL + jira.formatPath(path)
	headers := map[string]string{
		"Accept": "application/json",
	}
	return jira.http.GetWithAuth(ctx, url, headers, jira.userName, jira.userPwd)
}

func (jira *JiraTool) post(ctx context.Context, path, body string) ([]byte, error) {
	url := jira.restURL + jira.formatPath(path)
	headers := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
	return jira.http.PostWithAuth(ctx, url, headers, body, jira.userName, jira.userPwd)
}
