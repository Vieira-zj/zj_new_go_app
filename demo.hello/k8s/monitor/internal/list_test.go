package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	k8spkg "demo.hello/k8s/client/pkg"
)

var (
	listerTest *Lister
)

func init() {
	ns := "k8s-test"
	client, err := k8spkg.CreateK8sClientLocalDefault()
	if err != nil {
		panic(err)
	}
	watcher := NewWatcher(client, ns, 15)
	listerTest = NewLister(client, watcher, ns)
}

func TestGetAllPodInfos(t *testing.T) {
	infos, err := listerTest.GetAllPodInfosByRaw(context.Background())
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
