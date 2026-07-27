package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"wb_test_task/configs"
	"wb_test_task/internal/order"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/segmentio/kafka-go"
)

type OrderStore interface {
	Save(o order.Order) error
	PutCache(o *order.Order)
}

type Consumer struct {
	brokers         []string
	topic           string
	groupID         string
	orderRepository OrderStore
	validate        *validator.Validate
}

func NewConsumer(config *configs.Config, orderRepository OrderStore) *Consumer {
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
			continue
		}

		commit, procErr := c.processMessage(message.Value)
		if procErr != nil {
			log.Println("order processing failed", "error", procErr, "partition", message.Partition, "offset", message.Offset)
		}
		if !commit {

			continue
		}
		if err := reader.CommitMessages(ctx, message); err != nil {
			log.Println("kafka offset commit failed", "error", err)
			continue
		}
		if procErr == nil {
			log.Println("order processed")
		}
	}
}

func (c *Consumer) processMessage(payload []byte) (commit bool, procErr error) {
	var ord order.Order
	if err := json.Unmarshal(payload, &ord); err != nil {
		return true, fmt.Errorf("invalid message format: %w", err)
	}

	if err := c.validate.Struct(ord); err != nil {
		return true, fmt.Errorf("validation failed for order %s: %w", ord.OrderUID, describeValidationErr(err))
	}

	if err := c.orderRepository.Save(ord); err != nil {
		if isDuplicateKeyError(err) {
			c.orderRepository.PutCache(&ord)
			return true, nil
		}
		return false, fmt.Errorf("db write failed for order %s: %w", ord.OrderUID, err)
	}

	c.orderRepository.PutCache(&ord)
	return true, nil
}

func describeValidationErr(err error) error {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return err
	}
	parts := make([]string, 0, len(validationErrors))
	for _, e := range validationErrors {
		parts = append(parts, "field "+e.Field()+" failed on tag '"+e.Tag()+"' with value '"+e.Param()+"'")
	}
	return errors.New(strings.Join(parts, ";"))
}

func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}
