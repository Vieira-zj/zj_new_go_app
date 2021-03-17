package demos

import (
	"fmt"
	"testing"
	"time"
)

// runner
type runner struct {
	Status string // running, stop
	Events *events
}

func (r *runner) start() {
	r.Status = "running"
	time.Sleep(time.Second)
	r.Events.onStart.fire(r)
}

func (r *runner) stop() {
	r.Status = "stop"
	r.Events.onStop.fire(r)
}

func (r *runner) getStatus() string {
	return r.Status
}

// events
type events struct {
	onStart *eventHook
	onStop  *eventHook
}

func newEvents() *events {
	return &events{
		onStart: newEventHook(),
		onStop:  newEventHook(),
	}
}

// eventHook
type eventHook struct {
	Handlers map[string]handlerFunc
}

func newEventHook() *eventHook {
	return &eventHook{
		Handlers: make(map[string]handlerFunc),
	}
}

func (hook *eventHook) register(h *handler) {
	hook.Handlers[h.Name] = h.Fn
}

func (hook *eventHook) unregister(h *handler) {
	delete(hook.Handlers, h.Name)
}

func (hook *eventHook) fire(r *runner) {
	for _, fn := range hook.Handlers {
		if fn != nil {
			fn(r)
		}
	}
}

// handler
type handlerFunc func(ctx *runner) error

type handler struct {
	Name string
	Fn   handlerFunc
}

func TestEventHook(t *testing.T) {
	startHandler := &handler{
		Name: "start",
		Fn: func(ctx *runner) error {
			fmt.Println("init env")
			fmt.Println("runner status:", ctx.Status)
			return nil
		},
	}

	stopHandler := &handler{
		Name: "stop",
		Fn: func(ctx *runner) error {
			fmt.Println("clearup env")
			fmt.Println("runner status:", ctx.Status)
			return nil
		},
	}

	events := newEvents()
	events.onStart.register(startHandler)
	events.onStop.register(stopHandler)
	defer func() {
		events.onStart.unregister(startHandler)
		events.onStop.unregister(stopHandler)
		fmt.Println("handlers count:", len(events.onStart.Handlers), len(events.onStop.Handlers))
	}()

	myRunner := runner{
		Events: events,
	}
	myRunner.start()
	myRunner.stop()
	fmt.Println("event demo done")
}
