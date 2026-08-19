package kafka

import (
	"context"
	"fmt"

	"gocommerce/services/order/internal/ports"

	"github.com/segmentio/kafka-go"
)

// EventPublisher is the Kafka implementation of ports.EventPublisher
type EventPublisher struct {
	writer *kafka.Writer
}

// NewEventPublisher creates a new Kafka event publisher
func NewEventPublisher(brokers []string) ports.EventPublisher {
	return &EventPublisher{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Balancer:               &kafka.Hash{}, // Use hash balancer for consistent partitioning by key
			AllowAutoTopicCreation: true,
			RequiredAcks:           kafka.RequireAll, // Wait for all replicas to acknowledge
			Async:                  false,            // Synchronous writes for reliability
		},
	}
}

// Publish publishes an event to Kafka
func (p *EventPublisher) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish message to kafka: %w", err)
	}

	return nil
}

// Close closes the Kafka writer
func (p *EventPublisher) Close() error {
	return p.writer.Close()
}
