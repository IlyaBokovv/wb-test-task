package consumer

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"
	"wb_test_task/configs"
	"wb_test_task/internal/order"

	"github.com/go-playground/validator/v10"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	brokers         []string
	topic           string
	groupID         string
	orderRepository *order.OrderRepository
	validate        *validator.Validate
}

func NewConsumer(config *configs.Config, orderRepository *order.OrderRepository) *Consumer {
	return &Consumer{
		brokers:         config.Kafka.Brokers,
		topic:           config.Kafka.Topic,
		groupID:         config.Kafka.GroupId,
		orderRepository: orderRepository,
		validate:        validator.New(),
	}
}

func (c *Consumer) StartConsuming(ctx context.Context) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.brokers,
		Topic:          c.topic,
		GroupID:        c.groupID,
		MinBytes:       1,
		MaxBytes:       10 * 1024 * 1024,
		CommitInterval: 0,
	})
	defer reader.Close()
	log.Println("kafka consumer started", "topic", c.topic, "brokers", c.brokers)
	for {
		message, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Println("kafka receive failed", "error", err)
			if !retry(ctx) {
				return
			}
			continue
		}
		var order order.Order
		decodeErr := json.Unmarshal(message.Value, &order)
		if decodeErr != nil {
			log.Println("invalid kafka message format, message ignored", "error", decodeErr, "partition", message.Partition, "offset", message.Offset)
			if err := reader.CommitMessages(ctx, message); err != nil {
				log.Println("failed to commit invalid message", "error", err)
			}
			continue
		}

		if validationErr := c.validate.Struct(order); validationErr != nil {
			var validationErrors []string
			if errs, ok := validationErr.(validator.ValidationErrors); ok {
				for _, err := range errs {
					validationErrors = append(validationErrors,
						"field "+err.Field()+" failed on tag '"+err.Tag()+"' with value '"+err.Param()+"'")
				}
			} else {
				validationErrors = append(validationErrors, validationErr.Error())
			}

			if err := reader.CommitMessages(ctx, message); err != nil {
				log.Println("Kafka offset commit failed", "error", err)
				continue
			}
			log.Println("order validation failed, message ignored",
				"errors", strings.Join(validationErrors, "; "),
				"partition", message.Partition,
				"offset", message.Offset,
				"order_uid", order.OrderUID)
			continue
		}
		if err := c.orderRepository.Save(order); err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				log.Println("order with the same id already exists in db", "order_uid", order.OrderUID)
			} else {
				log.Println("database write failed, message remains uncommitted", "order_uid", order.OrderUID, "error", err)
				if !retry(ctx) {
					return
				}
				continue
			}
		}
		c.orderRepository.Cache.Put(&order)
		if err := reader.CommitMessages(ctx, message); err != nil {
			log.Println("Kafka offset commit failed", "error", err)
			continue
		}
		log.Println("order processed", "order_uid", order.OrderUID)
	}
}

func retry(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(time.Second):
		return true
	}
}
