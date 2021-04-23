package pkg

const (
	isDebug   = false
	expired   = 10 * 60 // seconds
	queueSize = 30
	mapSize   = 20
)

var (
	jira = NewJiraTool()
	git  = NewGitlabTool()
)
