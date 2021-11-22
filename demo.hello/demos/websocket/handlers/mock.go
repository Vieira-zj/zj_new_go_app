package handlers

import (
	"context"
	"fmt"
	"time"

	"demo.hello/utils"
)

var (
	echoChannel = "OnEchoMessage"
	eventBusSvr *utils.EventBusServer
)

// CreateEventBusServer .
func CreateEventBusServer() *utils.EventBusServer {
	eventBusSvr = utils.NewEventBusServer(16, 0)
	return eventBusSvr
}

// MockMessage .
func MockMessage(ctx context.Context) {
	fmt.Println("[mock]: start")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("[mock]: cancel")
			return
		default:
		}

		message, err := utils.GetRandString(8)
		if err != nil {
			message = err.Error()
		}
		if err := eventBusSvr.Publish(echoChannel, message); err != nil {
			fmt.Println("[mock] publish message error:", err)
		}
		time.Sleep(time.Second)
	}
}
