package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"demo.hello/utils"
)

const (
	defaultUser = "jin.zheng"
)

var (
	notify               *MatterMostNotify
	matterMostNotifyOnce sync.Once
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

// MustSendMessageToDefaultUser sends notify message without return a error.
func (notify *MatterMostNotify) MustSendMessageToDefaultUser(ctx context.Context, text string) {
	if err := notify.SendMessageToUser(ctx, defaultUser, text); err != nil {
		log.Println("MustSendMessageToDefaultUser error:", err)
	}
}

// SendMessageToDefaultUser .
func (notify *MatterMostNotify) SendMessageToDefaultUser(ctx context.Context, text string) error {
	return notify.SendMessageToUser(ctx, defaultUser, text)
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
