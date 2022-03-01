package server

import (
	"sync"

	"demo.hello/issuescache/pkg"
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
	// locker for TreeMap. TODO: 单锁, 写入性能问题. 使用 sharding map 代替
	locker = new(sync.RWMutex)
	// StoreCancelMap global cancel funcs, only for issues tree v1.
	// StoreCancelMap map[string]context.CancelFunc = make(map[string]context.CancelFunc)

	jira = pkg.NewJiraTool()
)
