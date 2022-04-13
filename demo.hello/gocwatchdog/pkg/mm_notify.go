package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"demo.hello/utils"
)

// MMMessage .
type MMMessage struct {
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
}

// MatterMostNotify .
type MatterMostNotify struct {
	baseURL string
	token   string
	channel string
	client  *utils.HTTPUtils
}

var (
	notify               *MatterMostNotify
	matterMostNotifyOnce sync.Once
)

// NewMatterMostNotify .
func NewMatterMostNotify() *MatterMostNotify {
	matterMostNotifyOnce.Do(func() {
		notify = &MatterMostNotify{
			baseURL: getParamFromEnv("MM_URL"),
			token:   getParamFromEnv("MM_TOKEN"),
			channel: getParamFromEnv("MM_CHANNEL"),
			client:  utils.NewDefaultHTTPUtils(),
		}
	})
	return notify
}

// SendMessageToDefaultUser .
func (notify *MatterMostNotify) SendMessageToDefaultUser(ctx context.Context, text string) error {
	return notify.SendMessageToUser(ctx, "jin.zheng", text)
}

// SendMessageToUser .
func (notify *MatterMostNotify) SendMessageToUser(ctx context.Context, user, text string) error {
	if len(user) > 0 {
		text = fmt.Sprintf("@%s %s", user, text)
	}
	return notify.SendMessage(ctx, text)
}

// SendMessage .
func (notify *MatterMostNotify) SendMessage(ctx context.Context, text string) error {
	message := MMMessage{
		ChannelID: notify.channel,
		Message:   text,
	}
	b, err := json.Marshal(&message)
	if err != nil {
		return fmt.Errorf("SendMessage json marshal failed: %w", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + notify.token,
	}
	if _, err = notify.client.Post(ctx, notify.baseURL+"/posts", headers, string(b)); err != nil {
		return fmt.Errorf("SendMessage post message failed: %w", err)
	}
	return nil
}
