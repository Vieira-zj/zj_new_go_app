package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"demo.hello/utils"
)

// MMMessage .
type MMMessage struct {
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
}

// MatterMost .
type MatterMost struct {
	baseURL string
	token   string
	channel string
	client  *utils.HTTPUtils
}

// NewMatterMost create an instance of MatterMost.
func NewMatterMost() (*MatterMost, error) {
	url := os.Getenv("MM_URL")
	if len(url) == 0 {
		return nil, errors.New("env var [MM_URL] is empty")
	}
	token := os.Getenv("MM_TOKEN")
	if len(token) == 0 {
		return nil, errors.New("env var [MM_TOKEN] is empty")
	}
	channel := os.Getenv("MM_CHANNEL")
	if len(channel) == 0 {
		return nil, errors.New("env var [MM_CHANNEL] is empty")
	}

	client := utils.NewDefaultHTTPUtils()
	return &MatterMost{
		baseURL: url,
		token:   token,
		channel: channel,
		client:  client,
	}, nil
}

// SendMessageToUser send message to given channel and At specified user.
func (mm *MatterMost) SendMessageToUser(ctx context.Context, user, text string) error {
	return mm.SendMessage(ctx, fmt.Sprintf("@%s %s", user, text))
}

// SendMessage send message to given channel.
func (mm *MatterMost) SendMessage(ctx context.Context, text string) error {
	headers := make(map[string]string, 1)
	headers["Authorization"] = "Bearer " + mm.token

	message := MMMessage{
		ChannelID: mm.channel,
		Message:   text,
	}
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	if _, err = mm.client.Post(ctx, mm.baseURL+"posts", headers, string(b)); err != nil {
		return err
	}
	return nil
}
