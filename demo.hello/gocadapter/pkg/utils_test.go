package pkg

import (
	"context"
	"fmt"
	"testing"
)

func TestIsAttachServerOK(t *testing.T) {
	host := "http://127.0.0.1:51025"
	ok := IsAttachServerOK(context.Background(), host)
	fmt.Println("service ok:", ok)
}

func TestRemoveUnhealthServicesFromGocSvrList(t *testing.T) {
	if err := RemoveUnhealthServicesFromGocSvrList(context.Background(), localHost); err != nil {
		t.Fatal(err)
	}
	fmt.Println("done")
}
