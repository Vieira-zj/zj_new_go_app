package dbr

import (
	"context"
	"log"
	"time"

	"github.com/gocraft/dbr/v2"
)

type MyEventReceiver struct{}

func (e *MyEventReceiver) Event(eventName string) {
	const tag = "MyEventReceiver.Event"
	log.Printf("[%s] event name: %s", tag, eventName)
}

func (e *MyEventReceiver) EventKv(eventName string, kvs map[string]string) {
	const tag = "MyEventReceiver.EventKv"
	e.Event(eventName)
	log.Printf("[%s] kvs: %+v", tag, kvs)
}

func (e *MyEventReceiver) EventErr(eventName string, err error) error {
	const tag = "MyEventReceiver.EventErr"
	log.Printf("[%s] event name: %s", tag, eventName)
	log.Printf("[%s] event error: %v", tag, err)
	return err
}

func (e *MyEventReceiver) EventErrKv(eventName string, err error, kvs map[string]string) error {
	const tag = "MyEventReceiver.EventErrKv"
	err = e.EventErr(eventName, err)
	log.Printf("[%s] kvs: %+v", tag, kvs)
	return err
}

func (e *MyEventReceiver) Timing(eventName string, nanoseconds int64) {
	const tag = "MyEventReceiver.Timing"
	nano := time.Duration(nanoseconds) * time.Nanosecond
	log.Printf("[%s] event name: %s", tag, eventName)
	log.Printf("[%s] timing: %+d milli secs", tag, nano.Milliseconds())
}

func (e *MyEventReceiver) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {
	const tag = "MyEventReceiver.TimingKv"
	e.Timing(eventName, nanoseconds)
	log.Printf("[%s] kvs: %+v", tag, kvs)
}

type MyKey string

var (
	testKey = MyKey("my-key-test")
	mockKey = MyKey("my-key-mock")
)

func (e *MyEventReceiver) SpanStart(ctx context.Context, eventName, query string) context.Context {
	const tag = "MyEventReceiver.SpanStart"
	log.Printf("[%s] event name: %s", tag, eventName)
	log.Printf("[%s] query: %s", tag, query)

	ctx = context.WithValue(ctx, testKey, "my-value-test")
	return ctx
}

func (e *MyEventReceiver) SpanError(ctx context.Context, err error) {
	const tag = "MyEventReceiver.SpanError"
	log.Printf("[%s] event error: %v", tag, err)
	log.Printf("[%s] custom context value: %s", tag, ctx.Value(testKey))
}

func (e *MyEventReceiver) SpanFinish(ctx context.Context) {
	const tag = "MyEventReceiver.SpanFinish"
	if ctx != nil {
		log.Printf("[%s] custom context value: %s", tag, ctx.Value(testKey))
	} else {
		log.Println("context is nil")
	}
}

// MockEventReceiver
//
// Need to update dbr source code to support dry run mode. For example, set "ExecMode:dryrun" in context.
//

type MockEventReceiver struct {
	*dbr.NullEventReceiver
}

func (e *MockEventReceiver) SpanStart(ctx context.Context, eventName, query string) context.Context {
	ctx = context.WithValue(ctx, mockKey, "mock-value")
	return ctx
}

func (e *MockEventReceiver) SpanFinish(ctx context.Context) {
	log.Printf("custom context value: %s", ctx.Value(mockKey))
}
