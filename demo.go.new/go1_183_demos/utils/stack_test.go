package utils_test

import (
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"demo.apps/utils"
)

var _allDone chan struct{}

func waitForDone() {
	<-_allDone
}

func TestStackAll(t *testing.T) {
	_allDone = make(chan struct{})
	defer close(_allDone)

	for i := 0; i < 5; i++ {
		go waitForDone()
	}

	cur := utils.Current()
	got := utils.All()

	for isBackgroundRunning(cur, got) {
		t.Log("wait until the background stacks are not runnable/running")
		runtime.Gosched()
		got = utils.All()
		time.Sleep(time.Second)
	}

	t.Log("all goroutines count:", len(got))
	t.Log("current goroutine:", cur.String())

	sort.Sort(byGoroutineID(got))
	for _, s := range got {
		t.Log("goroutine:", s.String())
	}
}

type byGoroutineID []utils.Stack

func (ss byGoroutineID) Len() int           { return len(ss) }
func (ss byGoroutineID) Less(i, j int) bool { return ss[i].ID() < ss[j].ID() }
func (ss byGoroutineID) Swap(i, j int)      { ss[i], ss[j] = ss[j], ss[i] }

func isBackgroundRunning(cur utils.Stack, stacks []utils.Stack) bool {
	for _, s := range stacks {
		if cur.ID() == s.ID() {
			continue
		}

		if strings.Contains(s.State(), "run") {
			return true
		}
	}

	return false
}
