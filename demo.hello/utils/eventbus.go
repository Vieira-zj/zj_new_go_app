package utils

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

// Event .
type Event struct {
	ID      string
	Channel string
	Message []interface{}
}

// Callback .
type Callback struct {
	Name string
	Fn   func(...interface{})
}

// EventBusServer .
type EventBusServer struct {
	running  bool
	queue    chan *Event
	channels map[string][]Callback
	locker   *sync.Mutex
	goPool   *GoPool
}

var (
	once      sync.Once
	_eventbus *EventBusServer
)

// NewEventBusServer .
// if poolSize = 0, it creates new goroutine for each callback instread of using go pool.
func NewEventBusServer(poolSize, queueSize int) *EventBusServer {
	once.Do(func() {
		var goPool *GoPool
		if poolSize > 0 {
			goPool = NewGoPool(poolSize, 0, 8*time.Second)
		}
		_eventbus = &EventBusServer{
			running:  false,
			queue:    make(chan *Event, queueSize),
			channels: make(map[string][]Callback, 4),
			locker:   &sync.Mutex{},
			goPool:   goPool,
		}
		_eventbus.Start()
	})
	return _eventbus
}

// Start .
func (bus *EventBusServer) Start() {
	if bus.running {
		return
	}
	bus.running = true
	go bus.run()
}

// Stop .
func (bus *EventBusServer) Stop() {
	if !bus.running {
		return
	}
	bus.running = false
	close(bus.queue)

	if bus.goPool != nil {
		bus.goPool.Stop(8)
	}
}

func (bus *EventBusServer) run() {
	for event := range bus.queue {
		localEvt := event
		callbacks := bus.channels[event.Channel]
		for _, cb := range callbacks {
			localCb := cb
			if bus.goPool != nil {
				if err := bus.goPool.Submit(func() {
					// Note: use "local" in closure
					localCb.Fn(localEvt.Message...)
				}); err != nil {
					fmt.Println("[EventBusServer]: submit callback error:", err)
				}
			} else {
				go localCb.Fn(localEvt.Message...)
			}
		}
	}
}

// Publish .
func (bus *EventBusServer) Publish(channel string, message ...interface{}) (err error) {
	if !bus.running {
		err = fmt.Errorf("[EventBusServer]: eventbus is not running")
		return
	}
	if len(bus.queue) != 0 && len(bus.queue) == cap(bus.queue) {
		err = fmt.Errorf("[EventBusServer]: execeed max queue size: %d", cap(bus.queue))
		return
	}

	event := &Event{
		ID:      strconv.Itoa(time.Now().Nanosecond()),
		Channel: channel,
		Message: message,
	}
	bus.queue <- event
	return
}

// Register .
func (bus *EventBusServer) Register(channel string, cb Callback) (err error) {
	bus.locker.Lock()
	defer bus.locker.Unlock()

	if _, ok := bus.channels[channel]; !ok {
		bus.channels[channel] = make([]Callback, 0)
	}
	if bus.isCallbackExist(channel, cb) {
		return fmt.Errorf("[EventBusServer]: callback [%s] is exist in channel [%s]", cb.Name, channel)
	}
	bus.channels[channel] = append(bus.channels[channel], cb)
	return
}

// Unregister .
func (bus *EventBusServer) Unregister(channel string, callback Callback) {
	bus.locker.Lock()
	defer bus.locker.Unlock()

	callbacks, ok := bus.channels[channel]
	if !ok {
		return
	}
	for idx, cb := range callbacks {
		if callback.Name == cb.Name {
			bus.channels[channel] = append(callbacks[:idx], callbacks[idx+1:]...)
			return
		}
	}
}

// PrintInfo .
func (bus *EventBusServer) PrintInfo() {
	fmt.Println("eventbus info:")
	fmt.Println("queue size:", len(bus.queue))
	for key, cbs := range bus.channels {
		fmt.Printf("channel [%s]: ", key)
		for _, cb := range cbs {
			fmt.Printf("%s,", cb.Name)
		}
		fmt.Println()
	}
}

func (bus *EventBusServer) isCallbackExist(channel string, callback Callback) bool {
	callbacks := bus.channels[channel]
	for _, cb := range callbacks {
		if callback.Name == cb.Name {
			return true
		}
	}
	return false
}
