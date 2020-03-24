package adapters

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/pangpanglabs/goutils/kafka"
)

var EventMessagePublisher *OrderEventMessagePublisher

type OrderEventMessagePublisher struct {
	producer *kafka.Producer
}

func NewOrderEventMessagePublisher(kafkaConfig kafka.Config) (*OrderEventMessagePublisher, error) {
	producer, err := kafka.NewProducer(kafkaConfig.Brokers, kafkaConfig.Topic, func(c *sarama.Config) {
		c.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
		c.Producer.Compression = sarama.CompressionGZIP     // Compress messages
		c.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms
	})

	if err != nil {
		return nil, err
	}

	orderEventMessagePublisher := OrderEventMessagePublisher{
		producer: producer,
	}

	return &orderEventMessagePublisher, nil
}

func (publisher OrderEventMessagePublisher) Close() {
	publisher.producer.Close()
}

func (publisher OrderEventMessagePublisher) Publish(message interface{}, key string) error {
	if err := publisher.producer.SendWithKey(message, key); err != nil {
		return err
	}

	return nil
}
