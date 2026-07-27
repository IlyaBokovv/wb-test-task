package producer

import (
	"context"
	"flag"
	"os"
	"time"
	"wb_test_task/configs"

	"github.com/segmentio/kafka-go"
)

const CharacterSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

type Producer struct {
	SampleOrderBytes []byte
	Writer           *kafka.Writer
}

func NewProducer(config *configs.Config) *Producer {
	file := flag.String("file", "sample-order.json", "JSON file to publish")
	body, _ := os.ReadFile(*file)
	return &Producer{
		SampleOrderBytes: body,
		Writer: &kafka.Writer{
			Addr:         kafka.TCP(config.Kafka.Brokers...),
			Topic:        config.Kafka.Topic,
			RequiredAcks: kafka.RequireAll,
			BatchSize:    100,
			BatchTimeout: 100 * time.Millisecond,
		},
	}
}

func (p *Producer) ProduceMessage(message []byte) error {
	return p.Writer.WriteMessages(context.Background(),
		kafka.Message{Value: message})
}

func (p *Producer) Close() error {
	return p.Writer.Close()
}
