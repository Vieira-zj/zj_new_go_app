package dbr

import (
	"context"
	"testing"
)

func TestNewDBConn(t *testing.T) {
	conn := NewDBConn()
	sess := conn.NewSession(nil)

	if err := sess.PingContext(context.Background()); err != nil {
		t.Fatal(err)
	}

	var (
		ctx   = context.Background()
		names = []string{}
	)
	n, err := sess.Select("name").From("user").LoadContext(ctx, &names)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("affect row:", n)
	t.Log("users:", names)
}

func TestMyEventReceiver(t *testing.T) {
	// events: SpanStart => SpanFinish => TimingKv
	conn := NewDBConn()
	event := &MyEventReceiver{}
	sess := conn.NewSession(event)

	var (
		ctx   = context.Background()
		users []User
	)
	n, err := sess.Select("name", "age").From("user").LoadContext(ctx, &users)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("affect row:", n)
	for idx := range users {
		u := users[idx]
		t.Logf("user: name=%s,age=%d", u.Name, u.Age)
	}
}

func TestMyEventReceiverWhenError(t *testing.T) {
	// events: SpanStart => SpanError => EventErrKv => SpanFinish => TimingKv
	conn := NewDBConn()
	event := &MyEventReceiver{}
	sess := conn.NewSession(event)

	var (
		ctx   = context.Background()
		users []User
	)
	n, err := sess.Select("name", "age").From("users").LoadContext(ctx, &users)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("affect row:", n)
}

func TestMockEventReceiver(t *testing.T) {
	// TODO:
}
