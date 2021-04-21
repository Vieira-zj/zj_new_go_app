package pkg

var (
	jira    = NewJiraTool()
	git     = NewGitlabTool()
	isDebug = false
	expired = 10 * 60 // seconds
)
