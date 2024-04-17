package utils_test

import (
	"testing"

	"demo.apps/utils"
)

func TestGetLocalIPAddr(t *testing.T) {
	addr, err := utils.GetLocalIPAddr()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("local ip addr:", addr)

	addr, err = utils.GetLocalIPAddrByDial()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("local ip addr:", addr)
}
