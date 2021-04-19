package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// JiraIssue struct for a jira issue.
type JiraIssue struct {
	Key           string   `json:"key"`
	Summary       string   `json:"summary"`
	Type          string   `json:"type"`
	Status        string   `json:"status"`
	Labels        []string `json:"labels"`
	FixVersions   []string `json:"fixVersions"`
	ReleaseCycle  string   `json:"releaseCycle"`
	ReleaseStatus string   `json:"releaseStatus"`
	SuperIssues   []string `json:"superIssue"`
	SubIssues     []string `json:"subIssues"`
	MergeRequests []string `json:"mergeRequests"`
}

// PrintText prints issue data by text.
func (issue *JiraIssue) PrintText(prefix string) {
	labels := getPrintFieldFromSlice(issue.Labels)
	fixVersions := getPrintFieldFromSlice(issue.FixVersions)
	superIssues := getPrintFieldFromSlice(issue.SuperIssues)
	subIssues := getPrintFieldFromSlice(issue.SubIssues)
	fmt.Printf("%s[%s]: [type:%s],[status:%s],[labels:%s],[fixversion:%s],[relCycle:%s],[relStatus:%s],[supIssues:%s],[subIssues:%s]\n",
		prefix, issue.Key, issue.Type, issue.Status, labels, fixVersions, issue.ReleaseCycle, issue.ReleaseStatus, superIssues, subIssues)

	if false {
		for _, mr := range issue.MergeRequests {
			fmt.Printf("%s\t[mr:%s]\n", prefix, mr)
		}
	}
}

func getPrintFieldFromSlice(slice []string) string {
	line := "-"
	if len(slice) > 0 {
		line = strings.Join(slice, ",")
	}
	return line
}

/*
New a jira issue V2.
*/

// RespJiraIssue .
type RespJiraIssue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
		Type    struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
		Labels      []string `json:"labels"`
		FixVersions []struct {
			Name string `json:"name"`
		} `json:"fixVersions"`
		ReleaseCycle struct {
			Value string `json:"value"`
		} `json:"customfield_13700"`
		ReleaseStatus string `json:"customfield_13801"`
		IssueLinks    []struct {
			Type struct {
				Inward  string `json:"inward"`
				Outward string `json:"outward"`
			} `json:"type"`
			InwardIssue struct {
				Key string `json:"key"`
			}
			OutwardIssue struct {
				Key string `json:"key"`
			}
		} `json:"issuelinks"`
	} `json:"fields"`
}

// RespRemoteLink .
type RespRemoteLink struct {
	Object struct {
		URL string `json:"url"`
	} `json:"object"`
}

// NewJiraIssueV2 .
func NewJiraIssueV2(ctx context.Context, jira *JiraTool, issueID string) (*JiraIssue, error) {
	fields := []string{"key", "summary", "issuetype", "status", "labels", "fixVersions", "customfield_13700", "customfield_13801", "issuelinks"}
	resp, err := jira.GetIssue(ctx, issueID, fields)
	if err != nil {
		return nil, err
	}

	respJiraIssue := &RespJiraIssue{}
	err = json.Unmarshal(resp, respJiraIssue)
	if err != nil {
		return nil, err
	}

	issue := &JiraIssue{
		Key:           respJiraIssue.Key,
		Summary:       respJiraIssue.Fields.Summary,
		Type:          respJiraIssue.Fields.Type.Name,
		Status:        respJiraIssue.Fields.Status.Name,
		Labels:        respJiraIssue.Fields.Labels,
		ReleaseCycle:  respJiraIssue.Fields.ReleaseCycle.Value,
		ReleaseStatus: respJiraIssue.Fields.ReleaseStatus,
	}
	fixIssueType(issue)

	fixVersions := make([]string, 0, len(respJiraIssue.Fields.FixVersions))
	for _, ver := range respJiraIssue.Fields.FixVersions {
		fixVersions = append(fixVersions, ver.Name)
	}
	issue.FixVersions = fixVersions

	issueLinks := make([]string, 0)
	if issue.Type == "PMTask" || issue.Type == "Story" || issue.Type == "Epic" || issue.Type == "Release" {
		for _, link := range respJiraIssue.Fields.IssueLinks {
			if link.Type.Outward == "Contains" && len(link.OutwardIssue.Key) > 0 {
				issueLinks = append(issueLinks, link.OutwardIssue.Key)
			}
		}
	}
	issue.SubIssues = issueLinks

	issueLinks = make([]string, 0)
	if issue.Type == "Task" || issue.Type == "Story" {
		for _, link := range respJiraIssue.Fields.IssueLinks {
			if link.Type.Inward == "In Release" && len(link.InwardIssue.Key) > 0 {
				issueLinks = append(issueLinks, link.InwardIssue.Key)
			}
		}
	}
	issue.SuperIssues = issueLinks

	resp, err = jira.GetRemoteLink(ctx, issueID)
	if err != nil {
		return nil, err
	}

	remoteLinks := make([]RespRemoteLink, 0)
	err = json.Unmarshal(resp, &remoteLinks)
	if err != nil {
		return nil, err
	}
	links := make([]string, 0, len(remoteLinks))
	for _, link := range remoteLinks {
		URL := link.Object.URL
		if strings.Contains(URL, "merge_requests") {
			links = append(links, URL)
		}
	}
	issue.MergeRequests = links

	return issue, nil
}

/*
New a jira issue.
*/

// NewJiraIssue creates a jira issue instance.
func NewJiraIssue(ctx context.Context, jira *JiraTool, issueID string) (*JiraIssue, error) {
	fields := []string{"key", "summary", "issuetype", "status", "labels", "fixVersions", "customfield_13700", "customfield_13801", "issuelinks"}
	resp, err := jira.GetIssue(ctx, issueID, fields)
	if err != nil {
		return nil, err
	}

	issueMap := make(map[string]interface{})
	err = json.Unmarshal(resp, &issueMap)
	if err != nil {
		return nil, err
	}

	// issue data
	fieldsMap := issueMap["fields"].(map[string]interface{})
	issue := &JiraIssue{
		Key:           issueMap["key"].(string),
		Summary:       fieldsMap["summary"].(string),
		Type:          fieldsMap["issuetype"].(map[string]interface{})["name"].(string),
		Status:        fieldsMap["status"].(map[string]interface{})["name"].(string),
		Labels:        formatLabelsSlice(fieldsMap["labels"].([]interface{})),
		FixVersions:   formatFixVersionsSlice(fieldsMap["fixVersions"].([]interface{})),
		ReleaseCycle:  createReleaseCycle(fieldsMap),
		ReleaseStatus: createReleaseStatus(fieldsMap),
	}
	fixIssueType(issue)

	// issue links
	issueLinks := fieldsMap["issuelinks"].([]interface{})
	if issue.Type == "PMTask" || issue.Type == "Story" || issue.Type == "Epic" || issue.Type == "Release" {
		issue.SubIssues = getOutwardIssueLinks(issueLinks, "Contains")
	}
	if issue.Type == "Task" || issue.Type == "Story" {
		issue.SuperIssues = getInwardIssueLinks(issueLinks, "In Release")
	}

	// remote links
	resp, err = jira.GetRemoteLink(ctx, issueID)
	if err != nil {
		return nil, err
	}
	remoteLinks := make([]interface{}, 0)
	err = json.Unmarshal(resp, &remoteLinks)
	if err != nil {
		return nil, err
	}
	issue.MergeRequests = getRemoteLinks(remoteLinks)

	return issue, nil
}

func fixIssueType(issue *JiraIssue) {
	if issue.Type == "Task" {
		for _, v := range issue.Labels {
			if v == "PM-Task" {
				issue.Type = "PMTask"
			}
		}
	}
}

func createReleaseCycle(fieldsMap map[string]interface{}) string {
	relCycle, ok := fieldsMap["customfield_13700"].(map[string]interface{})
	if !ok {
		return "not_fill"
	}
	return relCycle["value"].(string)
}

func createReleaseStatus(fieldsMap map[string]interface{}) string {
	relStatus, ok := fieldsMap["customfield_13801"].(string)
	if !ok {
		return "not_fill"
	}
	return relStatus
}

func formatLabelsSlice(labels []interface{}) []string {
	out := make([]string, 0, len(labels))
	for _, v := range labels {
		out = append(out, v.(string))
	}
	return out
}

func formatFixVersionsSlice(versions []interface{}) []string {
	out := make([]string, 0, len(versions))
	for _, v := range versions {
		val := v.(map[string]interface{})["name"]
		out = append(out, val.(string))
	}
	return out
}

func getInwardIssueLinks(issueLinks []interface{}, linkType string) []string {
	keys := make([]string, 0, 10)
	for _, v := range issueLinks {
		val := v.(map[string]interface{})
		if val["type"].((map[string]interface{}))["inward"].(string) == linkType {
			if subIssue, ok := val["inwardIssue"].(map[string]interface{}); ok {
				keys = append(keys, subIssue["key"].(string))
			}
		}
	}
	return keys
}

func getOutwardIssueLinks(issueLinks []interface{}, linkType string) []string {
	keys := make([]string, 0, 10)
	for _, v := range issueLinks {
		val := v.(map[string]interface{})
		if val["type"].((map[string]interface{}))["outward"].(string) == linkType {
			if subIssue, ok := val["outwardIssue"].(map[string]interface{}); ok {
				keys = append(keys, subIssue["key"].(string))
			}
		}
	}
	return keys
}

func getRemoteLinks(remoteLinks []interface{}) []string {
	links := make([]string, 0, len(remoteLinks))
	for _, link := range remoteLinks {
		mr := link.(map[string]interface{})["object"]
		mrURL := mr.(map[string]interface{})["url"].(string)
		if strings.Contains(mrURL, "merge_requests") {
			links = append(links, mrURL)
		}
	}
	return links
}
