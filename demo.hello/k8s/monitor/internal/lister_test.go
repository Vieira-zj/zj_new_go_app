package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	k8spkg "demo.hello/k8s/client/pkg"
	"k8s.io/client-go/kubernetes"
)

var (
	err    error
	client *kubernetes.Clientset
)

func init() {
	client, err = k8spkg.CreateK8sClientLocalDefault()
	if err != nil {
		panic(err)
	}
}

func TestJSONMarshal(t *testing.T) {
	// json marshal keeps "\n" in json output string of "Log"
	status := PodStatus{
		Namespace: "default",
		Name:      "test-pod",
		Value:     "Running",
		Log:       "multiple lines:\nline one\nline two",
	}

	b, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("json:\n%s", b)
}

func TestChannelCap(t *testing.T) {
	ch := make(chan string, 10)
	ch <- "hello"
	fmt.Printf("channel: len=%d, cap=%d\n", len(ch), cap(ch))
	close(ch)
}

func TestIsPodOK(t *testing.T) {
	for _, status := range []string{"Running", "Error"} {
		fmt.Printf("status [%s]: %v\n", status, isPodOK(status))
	}
}

func TestGetAllPodInfosByGivenNamespace(t *testing.T) {
	ns := "k8s-test"
	watcher := NewWatcher(client, []string{ns}, 15, true)
	lister := NewLister(client, watcher, []string{ns})

	infos, err := lister.GetAllPodInfosByRaw(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := prettyPrintJSON(infos); err != nil {
		t.Fatal(err)
	}
}

func TestGetAllPodInfosByMultiNamespaces(t *testing.T) {
	ns := "k8s-test,default"
	namespaces := strings.Split(ns, ",")
	watcher := NewWatcher(client, namespaces, 15, true)
	lister := NewLister(client, watcher, namespaces)

	infos, err := lister.GetAllPodInfosByRaw(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := prettyPrintJSON(infos); err != nil {
		t.Fatal(err)
	}
}

func prettyPrintJSON(data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("pods status info:\n%s\n", b)
	return nil
}

func TestDistinctByMap(t *testing.T) {}
