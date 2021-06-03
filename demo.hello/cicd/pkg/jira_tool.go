package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"demo.hello/utils"
)

/*
jira rest api docs:
https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/
*/

var _jira *JiraTool

// JiraTool contains jira rest APIs.
type JiraTool struct {
	restURL  string
	userName string
	userPwd  string
	http     *utils.HTTPUtils
}

// NewJiraTool creates a JiraTool instance.
func NewJiraTool() *JiraTool {
	locker.Lock()
	defer locker.Unlock()
	if _jira != nil {
		return _jira
	}

	_jira := &JiraTool{
		restURL:  jiraHost,
		userName: jiraUserName,
		userPwd:  jiraUserPwd,
		http:     utils.NewHTTPUtils(true),
	}
	return _jira
}

// GetIssue returns an issue by id.
func (jira *JiraTool) GetIssue(ctx context.Context, issueID string, fields []string) ([]byte, error) {
	path := "issue/" + issueID
	if fields != nil && len(fields) > 0 {
		path = path + "?fields=" + strings.Join(fields, ",")
	}
	return jira.get(ctx, path)
}

// Search finds issues by jql.
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

// SearchIssues returns issue keys by jql search.
func (jira *JiraTool) SearchIssues(ctx context.Context, jql string) ([]string, error) {
	resp, err := jira.Search(ctx, jql, []string{"key"})
	if err != nil {
		return nil, err
	}
	return getIssueKeysFromJQLResults(resp)
}

// GetIssuesInEpic retruns stories linked to a epic.
func (jira *JiraTool) GetIssuesInEpic(ctx context.Context, epicID string) ([]string, error) {
	jql := fmt.Sprintf(`"Epic Link" = %s`, epicID)
	resp, err := jira.Search(ctx, jql, []string{"key"})
	if err != nil {
		return nil, err
	}
	return getIssueKeysFromJQLResults(resp)
}

// GetIssueLink returns issue links.
func (jira *JiraTool) GetIssueLink(ctx context.Context, issueID string) ([]byte, error) {
	return jira.GetIssue(ctx, issueID, []string{"issuelinks"})
}

// GetRemoteLink returns issue remote links.
func (jira *JiraTool) GetRemoteLink(ctx context.Context, issueID string) ([]byte, error) {
	path := fmt.Sprintf("issue/%s/remotelink", issueID)
	return jira.get(ctx, path)
}

func (jira *JiraTool) get(ctx context.Context, path string) ([]byte, error) {
	url := jira.restURL + formatPath(path)
	headers := map[string]string{
		"Accept": textAppJSON,
	}
	return jira.http.GetWithAuth(ctx, url, headers, jira.userName, jira.userPwd)
}

func (jira *JiraTool) post(ctx context.Context, path, body string) ([]byte, error) {
	url := jira.restURL + formatPath(path)
	headers := map[string]string{
		"Accept":       textAppJSON,
		"Content-Type": textAppJSON,
	}
	return jira.http.PostWithAuth(ctx, url, headers, body, jira.userName, jira.userPwd)
}

/*
Common
*/

func formatPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

func getIssueKeysFromJQLResults(resp []byte) ([]string, error) {
	respMap := make(map[string]interface{})
	err := json.Unmarshal(resp, &respMap)
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
