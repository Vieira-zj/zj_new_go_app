package pkg

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestIssueKeySort(t *testing.T) {
	input := "SPayment-226,Payment-54042,Payment-56630,SPayment-227,Payment-56632,SPayment-2585,SPayment-3315,SPayment-69"
	issueKeys := strings.Split(input, ",")
	fmt.Println("before:", strings.Join(issueKeys, "|"))
	sort.Sort(byIssueKey(issueKeys))
	fmt.Println("after:", strings.Join(issueKeys, "|"))
}
