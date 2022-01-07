package internal

import (
	"context"
	"encoding/json"
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
func NewMatterMost() *MatterMost {
	client := utils.NewDefaultHTTPUtils()
	return &MatterMost{
		baseURL: getParamFromEnv("MM_URL"),
		token:   getParamFromEnv("MM_TOKEN"),
		channel: getParamFromEnv("MM_CHANNEL"),
		client:  client,
	}
}

// SendMessageToUser send message to given channel and At specified user.
func (mm *MatterMost) SendMessageToUser(ctx context.Context, user, text string) error {
	if len(user) > 0 {
		return mm.SendMessage(ctx, fmt.Sprintf("@%s %s", user, text))
	}
	return mm.SendMessage(ctx, text)
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

	if _, err = mm.client.Post(ctx, mm.baseURL+"/posts", headers, string(b)); err != nil {
		return err
	}
	return nil
}

func getParamFromEnv(param string) string {
	value := os.Getenv(param)
	if len(value) == 0 {
		panic(fmt.Sprintf("env var [%s] is empty", param))
	}
	return value
}
