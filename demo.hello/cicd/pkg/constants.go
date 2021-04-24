package pkg

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
	jira = NewJiraTool()
	git  = NewGitlabTool()
)
