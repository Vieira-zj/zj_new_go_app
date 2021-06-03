package pkg

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestIssueKeySort(t *testing.T) {
	input := "SPPAY-226,AIRPAY-54042,AIRPAY-56630,SPPAY-227,AIRPAY-56632,SPPAY-2585,SPPAY-3315,SPPAY-69"
	issueKeys := strings.Split(input, ",")
	fmt.Println("before:", strings.Join(issueKeys, "|"))
	sort.Sort(byIssueKey(issueKeys))
	fmt.Println("after:", strings.Join(issueKeys, "|"))
}
