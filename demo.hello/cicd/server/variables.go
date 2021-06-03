package server

import (
	"sync"

	"demo.hello/cicd/pkg"
)

const (
	typeReleaseCycle = "ReleaseCycle"
	typeFixVersion   = "FixVersion"
	typeJQL          = "jql"

	searchTimeout   = 3
	newStoreTimeout = 20
)

var (
	// Parallel .
	Parallel = 10
	// TreeMap .
	TreeMap = make(map[string]pkg.Tree)
	// StoreCancelMap global cancel funcs, only for issues tree v1.
	// StoreCancelMap map[string]context.CancelFunc = make(map[string]context.CancelFunc)

	jira   = pkg.NewJiraTool()
	locker = new(sync.RWMutex)
)
