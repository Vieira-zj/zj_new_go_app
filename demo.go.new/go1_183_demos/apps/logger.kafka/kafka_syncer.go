package main

import (
	"log"

	"github.com/Shopify/sarama"
	"go.uber.org/zap/zapcore"
)

type kafkaWriteSyncer struct {
	topic          string
	producer       sarama.AsyncProducer
	fallbackSyncer zapcore.WriteSyncer
}

// 第一个参数为sarama.AsyncProducer接口类型，目的是为了可以利用sarama提供的mock测试包
func NewKafkaSyncer(producer sarama.AsyncProducer, topic string, fallbackWs zapcore.WriteSyncer) zapcore.WriteSyncer {
	w := &kafkaWriteSyncer{
		producer:       producer,
		topic:          topic,
		fallbackSyncer: fallbackWs,
	}

	// 配置 config.Producer.Return.Errors = true
	// 处理写入失败的日志数据
	go func() {
		for e := range producer.Errors() {
			val, err := e.Msg.Value.Encode()
			if err != nil {
				log.Println("error from producer:", err.Error())
				continue
			}
			fallbackWs.Write(val)
		}
	}()
	return w
}

func (ws *kafkaWriteSyncer) Write(b []byte) (n int, err error) {
	// b会被zap重用，因此我们在扔给sarama channel之前需要将b copy一份，将副本发送给sarama
	copied := make([]byte, len(b))
	copy(copied, b)

	msg := &sarama.ProducerMessage{
		Topic: ws.topic,
		Value: sarama.ByteEncoder(copied),
	}

	ws.producer.Input() <- msg

	// 如果msg阻塞在Input channel上时，我们将日志写入fallbackSyncer. 问题：
	// 在压测中，我们发现大量日志都无法写入到kafka, 而是都写到了fallback syncer中。究其原因，我们在sarama的 async_producer.go 中看到：
	// input channel是一个unbuffered channel, 而从input channel读取消息的dispatcher goroutine也仅仅有一个。
	//
	// select {
	// case ws.producer.Input() <- msg:
	// default:
	// 	return ws.fallbackSyncer.Write(copied)
	// }

	return len(copied), nil
}

func (ws *kafkaWriteSyncer) Sync() error {
	ws.producer.AsyncClose()
	return ws.fallbackSyncer.Sync()
}

func NewKafkaAsyncProducer(addrs []string) (sarama.AsyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Errors = true
	return sarama.NewAsyncProducer(addrs, config)
}
