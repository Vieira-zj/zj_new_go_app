package demos

import (
	"sync"
	"testing"
	"time"
)

func TestBroadcaster(t *testing.T) {
	// b := NewBroadcaster1()
	// b := NewBroadcaster2()
	// b := NewBroadcaster3()
	b := NewBroadcaster4()
	// b := NewBroadcaster5()

	var wg sync.WaitGroup
	wg.Add(2)

	t.Run("submit tasks to broadcaster", func(t *testing.T) {
		b.Go(func() {
			t.Log("function 1 finished")
			wg.Done()
		})
		b.Go(func() {
			t.Log("function 2 finished")
			wg.Done()
		})

		time.Sleep(2 * time.Second)
		b.Broadcast()

		wg.Wait()
	})

	t.Run("submit task to exit broadcaster", func(t *testing.T) {
		b.Go(func() {
			t.Log("function 3 finished")
		})
		time.Sleep(time.Second)
	})

	t.Log("finished")
}
