package internal

import (
	"context"
	"testing"
	"time"
)

func TestSendMessage(t *testing.T) {
	mm := NewMatterMost()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	if err := mm.SendMessage(ctx, "hello world"); err != nil {
		t.Fatal(err)
	}
}

func TestSendMessageToUser(t *testing.T) {
	mm := NewMatterMost()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	if err := mm.SendMessageToUser(ctx, "jin.zheng", "hello world"); err != nil {
		t.Fatal(err)
	}
}
