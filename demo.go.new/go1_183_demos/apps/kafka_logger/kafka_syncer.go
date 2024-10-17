package kafkalogger

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

// 第一个参数为 sarama.AsyncProducer 接口类型，目的是为了可以利用 sarama 提供的 mock 测试包
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
	// b 会被 zap 重用，因此我们在扔给 sarama channel 之前需要将 b copy 一份，将副本发送给 sarama
	copied := make([]byte, len(b))
	copy(copied, b)

	msg := &sarama.ProducerMessage{
		Topic: ws.topic,
		Value: sarama.ByteEncoder(copied),
	}

	ws.producer.Input() <- msg

	// 如果 msg 阻塞在 input channel 上时，我们将日志写入 fallbackSyncer. 问题：
	// 在压测中，我们发现大量日志都无法写入到 kafka, 而是都写到了 fallback syncer 中。究其原因，我们在 sarama 的 async_producer.go 中看到：
	// input channel 是一个 unbuffered channel, 而从 input channel 读取消息的 dispatcher goroutine 也仅仅有一个。
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
