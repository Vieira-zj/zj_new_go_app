package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	k8spkg "demo.hello/k8s/client/pkg"
)

var listTestResource *k8spkg.Resource

func init() {
	client, err := k8spkg.CreateK8sClientLocalDefault()
	if err != nil {
		panic(err)
	}
	listTestResource = k8spkg.NewResource(client)
}

func TestGetAllPodInfos(t *testing.T) {
	ns := "k8s-test"
	infos, err := GetAllPodInfos(context.Background(), listTestResource, ns)
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
