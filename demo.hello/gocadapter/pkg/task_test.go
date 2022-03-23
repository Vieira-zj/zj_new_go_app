package pkg

import (
	"fmt"
	"testing"
)

func TestIsAttachServerOK(t *testing.T) {
	host := "http://127.0.0.1:51025"
	ok := isAttachServerOK(host)
	fmt.Println("service ok:", ok)
}

func TestRemoveUnhealthServicesFromGoc(t *testing.T) {
	if err := removeUnhealthServicesFromGoc(localHost); err != nil {
		t.Fatal(err)
	}
	fmt.Println("done")
}

func TestGetSrvCoverProcess(t *testing.T) {
	// TODO:
}
