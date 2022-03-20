package pkg

import (
	"fmt"
	"testing"
)

func TestIsAttachServerOK(t *testing.T) {
	host := "http://127.0.0.1:51025"
	ok := IsAttachServerOK(host)
	fmt.Println("service ok:", ok)
}

func TestRemoveUnhealthServicesFromGocSvrList(t *testing.T) {
	if err := removeUnhealthServicesFromGocSvrList(localHost); err != nil {
		t.Fatal(err)
	}
	fmt.Println("done")
}

func TestGetSrvCoverProcess(t *testing.T) {
	// TODO:
}
