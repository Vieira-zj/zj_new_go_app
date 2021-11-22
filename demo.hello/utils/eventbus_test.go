package utils

import (
	"fmt"
	"testing"
	"time"
)

var (
	eventbus       *EventBusServer
	onAddChannel   = "onAdd"
	onPrintChannel = "onPrint"
)

func init() {
	eventbus = NewEventBusServer(10, 3)
}

type calResult struct {
	val int
}

func (result *calResult) Add(a, b int) error {
	result.val = a + b
	return eventbus.Publish(onAddChannel, a, b)
}

func (result *calResult) Print() error {
	fmt.Println("result value:", result.val)
	return eventbus.Publish(onPrintChannel, result.val)
}

func newCalResult() *calResult {
	return &calResult{
		val: -1,
	}
}

func TestEventBus(t *testing.T) {
	defer func() {
		eventbus.Stop()
		eventbus.PrintInfo()
	}()

	// init eventbus
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
	eventbus.Register(onAddChannel, cbFoo)
	eventbus.Register(onAddChannel, cbBar)
	defer func() {
		eventbus.Unregister(onAddChannel, cbFoo)
		eventbus.Unregister(onAddChannel, cbBar)
	}()

	cbTest := Callback{
		Name: "test",
		Fn: func(val ...interface{}) {
			fmt.Println("[test] value:", val[0].(int))
		},
	}
	if err := eventbus.Register(onPrintChannel, cbTest); err != nil {
		t.Fatal(err)
	}
	defer eventbus.Unregister(onPrintChannel, cbTest)
	eventbus.PrintInfo()

	result := newCalResult()
	if err := result.Add(1, 2); err != nil {
		t.Fatal(err)
	}
	if err := result.Add(3, 4); err != nil {
		t.Fatal(err)
	}

	if err := result.Print(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
}
