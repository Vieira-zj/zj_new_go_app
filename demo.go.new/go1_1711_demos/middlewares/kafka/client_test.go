package main

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/Shopify/sarama"
)

// Admin

func TestListTopics(t *testing.T) {
	admin, err := GetLocalKafkaAdminForTest()
	if err != nil {
		t.Fatal("error while creating cluster admin: ", err.Error())
	}
	defer admin.Close()

	topics, err := admin.ListTopics()
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range topics {
		if k == "__consumer_offsets" {
			continue
		}
		t.Logf("topic=%s, total_partition=%d", k, v.NumPartitions)
	}
}

func TestCreateTopic(t *testing.T) {
	admin, err := GetLocalKafkaAdminForTest()
	if err != nil {
		t.Fatal("error while creating cluster admin: ", err.Error())
	}
	defer admin.Close()

	topic := "httpserver_important"
	if err := admin.CreateTopic(topic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false); err != nil {
		t.Fatal("error while creating topic: ", err.Error())
	}
	t.Log("create topic success:", topic)
}

func TestCreateLogTopic(t *testing.T) {
	admin, err := GetLocalKafkaAdminForTest()
	if err != nil {
		t.Fatal("error while creating cluster admin: ", err.Error())
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
		t.Fatal("error while creating topic: ", err.Error())
	}
	t.Log("create topic success:", topic)
}

func TestDeleteTopic(t *testing.T) {
	admin, err := GetLocalKafkaAdminForTest()
	if err != nil {
		t.Fatal("error while creating cluster admin: ", err.Error())
	}
	defer admin.Close()

	if err := admin.DeleteTopic("httpserver_access_log"); err != nil {
		t.Fatal("error while deleting topic: ", err.Error())
	}
	t.Log("delete topic success")
}

func TestListConsumerGroup(t *testing.T) {
	admin, err := GetLocalKafkaAdminForTest()
	if err != nil {
		t.Fatal("error while creating cluster admin: ", err.Error())
	}
	defer admin.Close()

	results, err := admin.ListConsumerGroups()
	if err != nil {
		t.Fatal(err)
	}
	groups := make([]string, 0, len(results))
	for name := range results {
		groups = append(groups, name)
	}

	descs, err := admin.DescribeConsumerGroups(groups)
	if err != nil {
		t.Fatal(err)
	}
	for _, desc := range descs {
		t.Log(desc.GroupId, desc.State)
	}
}

func TestDelConsumerGroup(t *testing.T) {
	admin, err := GetLocalKafkaAdminForTest()
	if err != nil {
		t.Fatal("error while creating cluster admin: ", err.Error())
	}
	defer admin.Close()

	if err = admin.DeleteConsumerGroup("consumer-group-test"); err != nil {
		t.Fatal(err)
	}
	t.Log("delete done")
}

// Consumer

func TestConsumeAll(t *testing.T) {
	master, err := GetLocalKafkaConsumerForTest()
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	msgCh, errCh, err := ConsumeAll(ctx, master)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for msg := range msgCh {
			t.Logf("get message: topic=%s, partition=%d, offset=%d, text=%s", msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
		}
		t.Log("consume exit")
	}()

	for err := range errCh {
		t.Log("consume error:", err.Topic, err.Partition, err.Error())
	}
	t.Log("consume done")
}
