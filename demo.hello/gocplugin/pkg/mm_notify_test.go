package pkg

import (
	"context"
	"testing"
)

func TestNotifySendMessage(t *testing.T) {
	// run: go test -timeout 3s -run ^TestNotifySendMessage$ demo.hello/gocplugin/pkg -v -count=1
	ctx := context.Background()
	notify := NewMatterMostNotify()
	if err := notify.SendMessageToUser(ctx, "jin.zheng", "Hello"); err != nil {
		t.Fatal(err)
	}
	if err := notify.SendMessage(ctx, "This is a notify test"); err != nil {
		t.Fatal(err)
	}
}
