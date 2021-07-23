package demos

import "time"

/*
runner
*/

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

/*
events
*/

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

/*
eventHook
*/

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

/*
handler
*/

type handlerFunc func(ctx *runner) error

type handler struct {
	Name string
	Fn   handlerFunc
}
