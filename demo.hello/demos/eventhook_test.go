package demos

import (
	"fmt"
	"testing"
)

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
