package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/Shopify/sarama"
)

// refer: https://github.com/Shopify/sarama/blob/9127f1c432c08b9a766fdb0931bfd0da80dddde0/examples/consumergroup/main.go

var (
	brokers  = ""
	version  = ""
	group    = ""
	topics   = ""
	assignor = ""
	oldest   = false
	verbose  = false
)

func init() {
	flag.StringVar(&brokers, "brokers", "localhost:9092", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&group, "group", "", "Kafka consumer group definition")
	flag.StringVar(&version, "version", "3.4.0", "Kafka cluster version")
	flag.StringVar(&topics, "topics", "", "Kafka topics to be consumed, as a comma separated list")
	flag.StringVar(&assignor, "assignor", "range", "Consumer group partition assignment strategy (range, roundrobin, sticky)")
	flag.BoolVar(&oldest, "oldest", true, "Kafka consumer consume initial offset from oldest")
	flag.BoolVar(&verbose, "verbose", true, "Sarama logging")
	flag.Parse()

	if len(topics) == 0 {
		panic("no topics given to be consumed, please set the -topics flag")
	}
	if len(group) == 0 {
		panic("no Kafka consumer group defined, please set the -group flag")
	}
}

// Start a consumer group client to loop consume topic messages.
// 1. Parallel is equal to partition number of topic
// 2. Mark consumed messages, and will continue from last offset

func main() {
	log.Println("starting a new Sarama consumer group")
	config := initConsumerGroupConfig()

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), group, config)
	if err != nil {
		log.Panicf("error creating consumer group client: %v", err)
	}
	defer client.Close()

	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := Consumer{
		ready: make(chan bool),
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			log.Println("consumer group client start consume")
			// `Consume` should be called inside an infinite loop, when a server-side rebalance happens,
			// the consumer session will need to be recreated to get the new claims
			if err := client.Consume(ctx, strings.Split(topics, ","), &consumer); err != nil {
				log.Panicf("error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("sarama consumer up and running!...")

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	keepRunning := true
	consumptionIsPaused := false
	for keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			log.Println("terminating: via signal")
			keepRunning = false
		case <-sigusr1:
			toggleConsumptionFlow(client, &consumptionIsPaused)
		}
	}

	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("error closing client: %v", err)
	}
}

func initConsumerGroupConfig() *sarama.Config {
	if verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	version, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		log.Panicf("error parsing Kafka version: %v", err)
	}

	config := sarama.NewConfig()
	config.ClientID = "go-kafka-consumer-group-for-test"
	config.Consumer.Return.Errors = true
	config.Version = version

	if oldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	switch assignor {
	case "sticky":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	case "roundrobin":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	case "range":
		config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	default:
		log.Panicf("unrecognized consumer group partition assignor: %s", assignor)
	}

	return config
}

func toggleConsumptionFlow(client sarama.ConsumerGroup, isPaused *bool) {
	if *isPaused {
		client.ResumeAll()
		log.Println("resuming consumption")
	} else {
		client.PauseAll()
		log.Println("pausing consumption")
	}
	*isPaused = !*isPaused
}

// Consumer

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	index atomic.Uint32
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	log.Println("consumer setup")
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("consumer clearup")
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	idx := consumer.index.Add(1)
	// NOTE: Do not move the code below to a goroutine. The `ConsumeClaim` itself is called within a goroutine.
	for {
		select {
		case message := <-claim.Messages():
			log.Printf("[%d] Message claimed: topic=%s,partition=%d,timestamp=%v,value=%s", idx, message.Topic, message.Partition, message.Timestamp, string(message.Value))
			session.MarkMessage(message, "")
		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance.
		case <-session.Context().Done():
			return nil
		}
	}
}
