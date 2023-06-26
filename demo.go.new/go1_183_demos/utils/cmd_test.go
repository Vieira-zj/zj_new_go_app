package utils_test

import (
	"testing"
	"time"

	"demo.apps/utils"
)

func TestRunShellCmd(t *testing.T) {
	for _, cmd := range []string{
		"sleep 3",
		"ls /tmp/test",
		"ls /tmp/not_exist",
	} {
		output, err := utils.RunShellCmd(cmd)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("cmd=%s, output: %s", cmd, output)
	}
	time.Sleep(time.Second)
}
