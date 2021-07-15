package utils

import (
	"fmt"
	"testing"
	"time"
)

var eventbus *EventBusServer

func init() {
	eventbus = NewEventBusServer(10)
	eventbus.Start()
}

type calResult struct {
	val     int
	channel string
}

func (result *calResult) Add(a, b int) error {
	result.val = a + b
	return eventbus.Publish(result.channel, a, b)
}

func newCalResult(channel string, cbs ...Callback) *calResult {
	for _, cb := range cbs {
		eventbus.Register(channel, cb)
	}
	return &calResult{
		val:     -1,
		channel: channel,
	}
}

func TestEventBus(t *testing.T) {
	cbFoo := Callback{
		Name: "foo",
		Fn: func(val ...interface{}) {
			fmt.Println("[foo] input values:", val[0].(int), val[1].(int))
		},
	}
	cbBar := Callback{
		Name: "bar",
		Fn: func(val ...interface{}) {
			fmt.Println("[foo] input args:", val[0].(int), val[1].(int))
		},
	}

	channel := "OnAdd"
	result := newCalResult(channel, cbFoo, cbBar)
	defer func() {
		eventbus.Unregister(channel, cbFoo)
		eventbus.Unregister(channel, cbBar)
		eventbus.Stop()
		eventbus.PrintInfo()
	}()

	if err := result.Add(1, 2); err != nil {
		t.Fatal(err)
	}
	if err := result.Add(3, 4); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
}
