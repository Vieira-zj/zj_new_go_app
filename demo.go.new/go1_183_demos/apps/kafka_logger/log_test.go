package kafkalogger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"go.uber.org/zap"
)

// refer: https://tonybai.com/2022/03/28/the-comparison-of-the-go-community-leading-kakfa-clients/
// git: https://github.com/bigwhite/experiments/tree/master/kafka-clients/zapkafka
//

func TestWriteFailWithKafkaSyncer(t *testing.T) {
	config := sarama.NewConfig()
	producer := mocks.NewAsyncProducer(t, config)

	var buf = make([]byte, 0, 256)
	w := bytes.NewBuffer(buf)
	w.Write([]byte("hello"))

	kafkaWriter := NewKafkaSyncer(producer, "topic-test", NewFileSyncer(w))
	logger := NewLogger(kafkaWriter, 0)

	// mock
	producer.ExpectInputAndFail(errors.New("produce error"))
	producer.ExpectInputAndFail(errors.New("produce error"))

	logger.Info("demo1", zap.String("status", "ok"))
	logger.Info("demo2", zap.String("status", "ok"))
	time.Sleep(time.Second)

	b := w.Bytes()
	if !bytes.Contains(b, []byte("demo1")) {
		t.Errorf("want true, got false")
	}
	if !bytes.Contains(b, []byte("demo2")) {
		t.Errorf("want true, got false")
	}
	t.Log("write to fallback:\n", string(b))

	if err := producer.Close(); err != nil {
		t.Fatal(err)
	}
	t.Log("test write fail KafkaSyncer done")
}

func TestWriteOKWithKafkaSyncer(t *testing.T) {
	config := sarama.NewConfig()
	producer := mocks.NewAsyncProducer(t, config)

	var buf = []byte{}
	w := bytes.NewBuffer(buf)

	topicName := "topic-test"
	kafkaWriter := NewKafkaSyncer(producer, topicName, NewFileSyncer(w))
	logger := NewLogger(kafkaWriter, 0)

	messageChecker := func(msg *sarama.ProducerMessage) error {
		b, err := msg.Value.Encode()
		if err != nil {
			return err
		}

		m := make(map[string]interface{})
		if err = json.Unmarshal(b, &m); err != nil {
			fmt.Printf("unmarshal error: %s\n", err)
			return err
		}

		if msg.Topic != topicName {
			return errors.New("invalid topic")
		}
		fmt.Printf("topic=%s, value=%s\n", msg.Topic, m)

		v, ok := m["msg"].(string)
		if !ok {
			err = errors.New("invalid msg")
			fmt.Printf("type assertion error: %s\n", err)
			return err
		}
		if v != "demo1" {
			err = errors.New("invalid msg value")
			fmt.Printf("type assertion error: %s\n", err)
			return err
		}

		v, ok = m["status"].(string)
		if !ok {
			err = errors.New("invalid status")
			fmt.Printf("type assert error")
			return err
		}
		if v != "ok" {
			return errors.New("invalid status value")
		}
		return nil
	}

	producer.ExpectInputWithMessageCheckerFunctionAndSucceed(mocks.MessageChecker(messageChecker))
	logger.Info("demo1", zap.String("status", "ok"))

	if err := producer.Close(); err != nil {
		t.Error(err)
	}
	if len(w.Bytes()) != 0 {
		t.Errorf("want 0, got %d", len(w.Bytes()))
	}
	t.Log("test write ok KafkaSyncer done")
}
