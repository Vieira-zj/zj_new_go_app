package pkg

import (
	"os"
)

const (
	textAppJSON = "application/json"

	issueTypeEpic    = "Epic"
	issueTypeStory   = "Story"
	issueTypePMTask  = "PMTask"
	issueTypeRelease = "Release"
	issueTypeTask    = "Task"
	issueTypeBug     = "Bug"

	isDebug   = false
	expired   = 10 * 60 // seconds
	queueSize = 30
	mapSize   = 20
)

var (
	jiraHost, jiraUserName, jiraUserPwd string
	gitlabHost, gitlabToken             string
)

func init() {
	jiraHost = os.Getenv("JIRA_REST_URL")
	jiraUserName = os.Getenv("JIRA_USER_NAME")
	jiraUserPwd = os.Getenv("JIRA_USER_PASSWORD")
	if len(jiraHost) == 0 || len(jiraUserName) == 0 || len(jiraUserPwd) == 0 {
		panic("env JIRA_REST_URL or JIRA_USER_NAME or JIRA_USER_PASSWORD is not set")
	}

	gitlabHost = os.Getenv("GITLAB_HOST")
	gitlabToken = os.Getenv("GITLAB_TOKEN")
	if len(gitlabHost) == 0 || len(gitlabToken) == 0 {
		panic("env GITLAB_HOST or GITLAB_TOKEN is not set")
	}
}
