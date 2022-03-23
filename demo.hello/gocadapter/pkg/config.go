package pkg

import "time"

const (
	shortWait = 3 * time.Second
	wait      = 5 * time.Second
	longWait  = 8 * time.Second
)

var (
	// WorkingRootDir .
	WorkingRootDir string
)

// ModuleToRepoMap .
var ModuleToRepoMap = map[string]string{
	"echoserver": "git@github.com:Vieira-zj/zj_new_go_app.git",
}
