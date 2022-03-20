package pkg

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	// ShortWait .
	ShortWait = 3 * time.Second
	// Wait .
	Wait = 5 * time.Second
	// LongWait .
	LongWait = 8 * time.Second
)

func getParamFromEnv(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		panic(fmt.Sprintf("Env var [%s] not set", key))
	}
	return value
}

func getSimpleNowDatetime() string {
	return time.Now().Format("20060102_150405")
}

func getBranchAndCommitFromSrvName(name string) (string, string) {
	items := strings.Split(name, "_")
	commitID := items[len(items)-1]
	branch := items[len(items)-2]
	return branch, commitID
}
