package pkg

import (
	"fmt"
	"strings"
)

/*
Print tree text.
*/

// GetIssuesTreeText .
func GetIssuesTreeText(tree Tree) string {
	outLines := make([]string, 0, 100)
	errLines := make([]string, 0, 10)
	errLines = append(errLines, "\n[Failed Tasks:]\n")

	outLines = append(outLines, "\n[Epic:]\n")
	for epicID := range tree.GetEpics() {
		epic, epicText := GetIssueAndMRsText(tree, epicID, "")
		if epic == nil {
			errLines = append(errLines, epicText)
			continue
		}
		outLines = append(outLines, epicText)
		for _, storyID := range epic.SubIssues {
			story, storyText := GetIssueAndMRsText(tree, storyID, "\t")
			if story == nil {
				errLines = append(errLines, storyText)
				continue
			}
			outLines = append(outLines, storyText)
			for _, issueID := range story.SubIssues {
				if issue, issueText := GetIssueAndMRsText(tree, issueID, "\t\t"); issue == nil {
					errLines = append(errLines, issueText)
				} else {
					outLines = append(outLines, issueText)
				}
			}
		}
		outLines = append(outLines, "\n")
	}

	outLines = append(outLines, "\n[Stories / PM-Tasks:]\n")
	for storyID := range tree.GetStories() {
		story, storyText := GetIssueAndMRsText(tree, storyID, "")
		if story == nil {
			errLines = append(errLines, storyText)
			continue
		}
		if isDulplicatedIssue(tree, story) {
			continue
		}
		outLines = append(outLines, storyText)
		for _, issueID := range story.SubIssues {
			if issue, issueText := GetIssueAndMRsText(tree, issueID, "\t"); issue == nil {
				errLines = append(errLines, issueText)
			} else {
				outLines = append(outLines, issueText)
			}
		}
		outLines = append(outLines, "\n")
	}

	outLines = append(outLines, "\n[Single Tasks / Bugs:]\n")
	for issueID := range tree.GetIssueStore().GetItems() {
		if issue, issueText := GetIssueAndMRsText(tree, issueID, ""); issue != nil {
			if (issue.Type == issueTypeTask || issue.Type == issueTypeBug) && !isDulplicatedIssue(tree, issue) {
				outLines = append(outLines, issueText)
				outLines = append(outLines, "\n")
			}
		} else {
			errLines = append(errLines, issueText)
		}
	}

	outLines = append(outLines, removeDulpicatedItem(errLines)...)
	return strings.Join(outLines, "")
}

// GetIssueAndMRsText .
func GetIssueAndMRsText(tree Tree, issueID string, prefix string) (*JiraIssue, string) {
	value, err := tree.GetIssueStore().Get(issueID)
	if err != nil {
		return nil, fmt.Sprintf("%sGet issue [%s] from store failed: %v\n", prefix, issueID, err)
	}
	issue := value.(*JiraIssue)
	if len(issue.Err) > 0 {
		return nil, fmt.Sprintf("%s[%s]: %s", prefix, issue.Key, issue.Err)
	}

	retlines := make([]string, 10)
	retlines = append(retlines, prefix+issue.ToText())
	retlines = append(retlines, getIssueMRsText(tree, issue, prefix+"\t"))
	return issue, strings.Join(retlines, "")
}

// GetIssuesTreeUsageText returns tree store usage.
func GetIssuesTreeUsageText(tree Tree) string {
	outLines := make([]string, 0, 4)
	outLines = append(outLines, "\nTree issue store usage:\n")
	outLines = append(outLines, tree.GetIssueStore().UsageToText())
	outLines = append(outLines, "Tree MR store usage:\n")
	outLines = append(outLines, tree.GetMRStore().UsageToText())
	return strings.Join(outLines, "")
}

// GetIssuesTreeSummaryText .
func GetIssuesTreeSummaryText(tree Tree) string {
	outLines := make([]string, 0, 4)
	outLines = append(outLines, fmt.Sprintf("\nTotal Epics: %d\n", len(tree.GetEpics())))
	outLines = append(outLines, fmt.Sprintf("Total Stories (PMTask): %d\n", len(tree.GetStories())))
	totalTasks := tree.GetIssueStore().Size() - len(tree.GetEpics()) - len(tree.GetStories())
	outLines = append(outLines, fmt.Sprintf("Total Tasks (include Bugs): %d\n", totalTasks))
	return strings.Join(outLines, "")
}

func getIssueMRsText(tree Tree, issue *JiraIssue, prefix string) string {
	outLines := make([]string, 5)
	dulplicatedMR := make(map[string]struct{}, 5)
	for _, mrURL := range issue.MergeRequests {
		if _, ok := dulplicatedMR[mrURL]; ok {
			continue
		}
		dulplicatedMR[mrURL] = struct{}{}

		value, err := tree.GetMRStore().Get(mrURL)
		if err != nil {
			line := fmt.Sprintf("%sGet MR [%s] from store failed: %v\n", prefix, mrURL, err)
			outLines = append(outLines, line)
			continue
		}
		mr := value.(*MergeRequest)
		if len(mr.Err) > 0 {
			outLines = append(outLines, fmt.Sprintf("%s[%s]: %s", prefix, mr.WebURL, mr.Err))
			continue
		}

		if mr.TargetBR == "master" {
			outLines = append(outLines, prefix+mr.ToText())
		}
	}
	return strings.Join(outLines, "")
}

func isDulplicatedIssue(tree Tree, issue *JiraIssue) bool {
	for _, supIssueID := range issue.SuperIssues {
		if _, err := tree.GetIssueStore().Get(supIssueID); err == nil {
			return true
		}
	}
	return false
}

/*
Print deploy repo text.
*/

// GetDeployReposText .
func GetDeployReposText(tree Tree) string {
	outLines := make([]string, 0, 50)
	outLines = append(outLines, "\n[Deploy Repos:]\n")
	errLines := make([]string, 0)
	errLines = append(errLines, "[Error Merge Requests:]\n")

	reposToMRS := getReposToMRs(tree)
outer:
	for repo, mrs := range reposToMRS {
		mrWebURLs := make([]string, 0)
		issueIDs := make([]string, 0)
		for _, mr := range mrs {
			if len(mr.Err) > 0 {
				errLines = append(errLines, fmt.Sprintf("[mr:%s]: %s", mr.WebURL, mr.Err))
				continue outer
			}
			mrWebURLs = append(mrWebURLs, mr.WebURL)
			issueIDs = append(issueIDs, mr.LinkedIssue)
		}

		stoiresIDs := make([]string, 0)
		issueIDs = removeDulpicatedItem(issueIDs)
		for _, issueID := range issueIDs {
			if item, err := tree.GetIssueStore().Get(issueID); err == nil {
				issue := item.(*JiraIssue)
				stoiresIDs = append(stoiresIDs, issue.SuperIssues...)
			} else {
				fmt.Printf("Failed get issue [%s] in store.\n", issueID)
				stoiresIDs = append(stoiresIDs, fmt.Sprintf("Get_Err_%s", issueID))
			}
		}

		outLines = append(outLines, fmt.Sprintf("[repo:%s]:\n", repo))
		outLine := fmt.Sprintf("\t[mrs:%s],[issues:%s],[stories:%s]\n\n",
			strings.Join(mrWebURLs, ","), strings.Join(issueIDs, ","), strings.Join(stoiresIDs, ","))
		outLines = append(outLines, outLine)
	}
	return strings.Join(outLines, "") + strings.Join(errLines, "")
}

// GetShortDeployReposText .
func GetShortDeployReposText(tree Tree) string {
	reposToMRS := getReposToMRs(tree)
	outLines := make([]string, 0, len(reposToMRS))
	outLines = append(outLines, fmt.Sprintf("\n[All Deploy Repos (total:%d):]\n", len(reposToMRS)))
	for repo := range reposToMRS {
		outLines = append(outLines, "\t"+repo+"\n")
	}
	return strings.Join(outLines, "")
}

func getReposToMRs(tree Tree) map[string][]*MergeRequest {
	allMRs := tree.GetMRStore().GetItems()
	reposToMRs := make(map[string][]*MergeRequest, len(allMRs)/2)
	for _, item := range allMRs {
		mr := item.(*MergeRequest)
		if mrs, ok := reposToMRs[mr.Repo]; ok {
			mrs = append(mrs, mr)
		} else {
			reposToMRs[mr.Repo] = []*MergeRequest{mr}
		}
	}
	return reposToMRs
}
