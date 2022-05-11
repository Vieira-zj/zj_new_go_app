package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"demo.hello/k8s/monitor/internal"
)

func TestJsonUnmarshal(t *testing.T) {
	type testLine struct {
		Total int      `json:"total"`
		Words []string `json:"words"`
	}

	line := testLine{}
	str := `{"total": 3, "words": ["one", "two", "three"]}`
	if err := json.Unmarshal([]byte(str), &line); err != nil {
		t.Fatal(err)
	}
	fmt.Println(line.Total, strings.Join(line.Words, "|"))

	newLine := testLine{}
	str = `{"total": 0}`
	if err := json.Unmarshal([]byte(str), &newLine); err != nil {
		t.Fatal(err)
	}
	fmt.Println(newLine.Total, len(newLine.Words), newLine.Words)
}

func TestFilterPods01(t *testing.T) {
	filter := &PodFilter{
		NameSpaces: []string{"ns1"},
		Names:      []string{"pod1", "pod2"},
	}

	pods := []*internal.PodStatus{
		{Namespace: "ns1", Name: "pod1x"},
		{Namespace: "ns1", Name: "pod2x"},
		{Namespace: "ns1", Name: "pod3"},
		{Namespace: "ns2", Name: "pod1"},
		{Namespace: "ns2", Name: "pod2"},
	}

	fmt.Println("filter pod:")
	for _, pod := range filterPods(pods, filter) {
		fmt.Printf("%s, %s\n", pod.Namespace, pod.Name)
	}
}

func TestFilterPods02(t *testing.T) {
	filter := &PodFilter{
		Names: []string{"pod1", "pod2"},
	}

	pods := []*internal.PodStatus{
		{Namespace: "ns1", Name: "pod1x"},
		{Namespace: "ns1", Name: "pod2x"},
		{Namespace: "ns1", Name: "pod3"},
		{Namespace: "ns2", Name: "pod1y"},
	}

	fmt.Println("filter pod:")
	for _, pod := range filterPods(pods, filter) {
		fmt.Printf("%s, %s\n", pod.Namespace, pod.Name)
	}
}
