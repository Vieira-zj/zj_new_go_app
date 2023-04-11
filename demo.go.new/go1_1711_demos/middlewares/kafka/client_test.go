package main

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/Shopify/sarama"
)

// Admin

func TestCreateTopic(t *testing.T) {
	admin, err := GetLocalKafkaAdminForTest()
	if err != nil {
		t.Fatal("Error while creating cluster admin: ", err.Error())
	}
	defer admin.Close()

	topic := "httpserver_important"
	if err := admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false); err != nil {
		t.Fatal("Error while creating topic: ", err.Error())
	}
	t.Log("create topic success:", topic)
}

func TestCreateLogTopic(t *testing.T) {
	admin, err := GetLocalKafkaAdminForTest()
	if err != nil {
		t.Fatal("Error while creating cluster admin: ", err.Error())
	}
	defer admin.Close()

	// topic config:
	// https://kafka.apache.org/documentation/#topicconfigs
	configMaxMessageBytes := strconv.Itoa(1024)            // 1k
	configRetentionBytes := strconv.Itoa(16 * 1024 * 1024) // 16m
	configRetentionMs := strconv.Itoa(3 * 60 * 1000)       // 3min

	topic := "httpserver_access_log"
	if err := admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     3,
		ReplicationFactor: 1,
		ConfigEntries: map[string]*string{
			"max.message.bytes": &configMaxMessageBytes,
			"retention.bytes":   &configRetentionBytes,
			"retention.ms":      &configRetentionMs,
		},
	}, false); err != nil {
		t.Fatal("Error while creating topic: ", err.Error())
	}
	t.Log("create topic success:", topic)
}

func TestDeleteTopic(t *testing.T) {
	admin, err := GetLocalKafkaAdminForTest()
	if err != nil {
		t.Fatal("Error while creating cluster admin: ", err.Error())
	}
	defer admin.Close()

	if err := admin.DeleteTopic("httpserver_access_log"); err != nil {
		t.Fatal("Error while deleting topic: ", err.Error())
	}
	t.Log("delete topic success")
}

func TestConsumeMessages(t *testing.T) {
	master, err := GetLocalKafkaConsumerForTest()
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()

	topics, err := master.Topics()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	msgCh, errCh, err := Consume(ctx, topics, master)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for msg := range msgCh {
			t.Logf("get message: topic=%s, partition=%d, offset=%d, text=%s", msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
		}
	}()

	for err := range errCh {
		t.Log("consume error:", err.Topic, err.Partition, err.Error())
	}
	t.Log("consume done")
}
