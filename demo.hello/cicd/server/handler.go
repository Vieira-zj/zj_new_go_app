package server

import (
	"net/http"
	"strconv"
	"sync"

	"demo.hello/cicd/pkg"
	"github.com/labstack/echo"
)

const (
	typeReleaseCycle = "ReleaseCycle"
	typeFixVersion   = "FixVersion"
	typeJQL          = "jql"
)

var (
	// Parallel .
	Parallel = 10
	// TreeMap .
	TreeMap = make(map[string]pkg.Tree)
	// StoreCancelMap global cancel funcs, only for issues tree v1.
	// StoreCancelMap map[string]context.CancelFunc = make(map[string]context.CancelFunc)

	count           = 0
	searchTimeout   = 3
	newStoreTimeout = 20
	locker          = new(sync.RWMutex)
	jira            = pkg.NewJiraTool()
)

// Index .
func Index(c echo.Context) error {
	count++
	return c.String(http.StatusOK, strconv.Itoa(count))
}

// Ping .
func Ping(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
