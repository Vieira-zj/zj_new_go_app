package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
)

func GetLocalKafkaConsumerForTest() (sarama.Consumer, error) {
	conf := sarama.NewConfig()
	conf.ClientID = "go-kafka-consumer-for-test"
	conf.Consumer.Return.Errors = true

	brokers := []string{"localhost:9092"}
	return sarama.NewConsumer(brokers, conf)
}

func GetLocalKafkaAdminForTest() (sarama.ClusterAdmin, error) {
	version, err := sarama.ParseKafkaVersion("3.4.0")
	if err != nil {
		return nil, err
	}

	brokerList := []string{"localhost:9092"}
	config := sarama.NewConfig()
	config.Version = version
	return sarama.NewClusterAdmin(brokerList, config)
}

// ConsumeAll consumes queue messages for all topics and partitions.
func ConsumeAll(ctx context.Context, master sarama.Consumer) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError, error) {
	topics, err := master.Topics()
	if err != nil {
		return nil, nil, err
	}
	log.Println("available topics:", topics)

	closeOnce := &sync.Once{}
	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)
	for _, topic := range topics {
		if strings.Contains(topic, "__consumer_offsets") {
			continue
		}

		partitions, err := master.Partitions(topic)
		if err != nil {
			return nil, nil, fmt.Errorf("get topic partitions error: %v", err)
		}

		for _, partition := range partitions {
			consumer, err := master.ConsumePartition(topic, partition, sarama.OffsetOldest)
			if nil != err {
				return nil, nil, fmt.Errorf("consumer partition error: topic=%s, partition=%d, error=%v", topic, partition, err)
			}

			log.Printf("start consuming: topic=%s, partition=%d", topic, partition)
			go func(ctx context.Context, topic string, partition int32, consumer sarama.PartitionConsumer) {
				for {
					select {
					case <-ctx.Done():
						log.Printf("consumer exit: topic=%s, partition=%d, reason=%s", topic, partition, ctx.Err())
						closeOnce.Do(func() {
							close(consumers)
							close(errors)
						})
						return
					case consumerErr := <-consumer.Errors():
						errors <- consumerErr
					case msg := <-consumer.Messages():
						consumers <- msg
					}
				}
			}(ctx, topic, partition, consumer)
		}
	}

	return consumers, errors, nil
}
