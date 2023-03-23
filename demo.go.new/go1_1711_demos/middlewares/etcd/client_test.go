package etcd

import (
	"context"
	"testing"
)

func TestNewEtcdClient(t *testing.T) {
	client := NewEtcdClient()
	status, err := client.Status(context.Background(), localEtcdEndpoint)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("version:", status.Version)
}
