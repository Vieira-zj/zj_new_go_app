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
	brokerList := []string{"localhost:9092"}
	config := sarama.NewConfig()
	config.Version = sarama.V2_0_0_0
	return sarama.NewClusterAdmin(brokerList, config)
}

func Consume(ctx context.Context, topics []string, master sarama.Consumer) (chan *sarama.ConsumerMessage, chan *sarama.ConsumerError, error) {
	closeOnce := &sync.Once{}
	consumers := make(chan *sarama.ConsumerMessage)
	errors := make(chan *sarama.ConsumerError)

	for _, topic := range topics {
		if strings.Contains(topic, "__consumer_offsets") {
			continue
		}

		partitions, err := master.Partitions(topic)
		if err != nil {
			return nil, nil, fmt.Errorf("get partitions error: %v", err)
		}
		// this only consumes partition no 1, you would probably want to consume all partitions
		consumer, err := master.ConsumePartition(topic, partitions[0], sarama.OffsetOldest)
		if nil != err {
			return nil, nil, fmt.Errorf("consumer partition error: topic=%s, partitions=%d, error=%v", topic, partitions[0], err)
		}

		log.Println("start consuming topic:", topic)
		go func(ctx context.Context, topic string, consumer sarama.PartitionConsumer) {
			for {
				select {
				case <-ctx.Done():
					log.Printf("consumer for topic [%s] exit: %s", topic, ctx.Err())
					closeOnce.Do(func() {
						close(consumers)
						close(errors)
					})
					return
				case consumerError := <-consumer.Errors():
					errors <- consumerError
				case msg := <-consumer.Messages():
					consumers <- msg
				}
			}
		}(ctx, topic, consumer)
	}

	return consumers, errors, nil
}
