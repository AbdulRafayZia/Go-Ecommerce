package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"

	"github.com/segmentio/kafka-go"

	"gocommerce/pkg/logger"
)

// EventPublisher publishes payment events to Kafka
type EventPublisher struct {
	writers map[string]*kafka.Writer
	logger  *logger.Logger
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(brokers []string, log *logger.Logger) *EventPublisher {
	writers := make(map[string]*kafka.Writer)

	// Create writers for each topic
	topics := []string{
		TopicPaymentAuthorized,
		TopicPaymentCaptured,
		TopicPaymentFailed,
		TopicPaymentCancelled,
		TopicPaymentRefunded,
	}

	for _, topic := range topics {
		writers[topic] = kafka.NewWriter(kafka.WriterConfig{
			Brokers:  brokers,
			Topic:    topic,
			Balancer: &kafka.Hash{}, // Use hash balancer for consistent partitioning
			// Synchronous writes for reliability
			Async:        false,
			RequiredAcks: int(kafka.RequireAll), // Wait for all replicas
		})
	}

	return &EventPublisher{
		writers: writers,
		logger:  log,
	}
}

// PublishPaymentEvent publishes a payment event to the appropriate topic
func (p *EventPublisher) PublishPaymentEvent(ctx context.Context, event *PaymentEvent) error {
	// Determine topic based on event type
	topic := p.getTopicForEventType(event.EventType)

	writer, exists := p.writers[topic]
	if !exists {
		return fmt.Errorf("no writer found for topic: %s", topic)
	}

	// Serialize event to JSON
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Use payment ID as key for consistent partitioning
	key := []byte(event.PaymentID)

	message := kafka.Message{
		Key:   key,
		Value: value,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(event.EventType)},
			{Key: "payment_id", Value: []byte(event.PaymentID)},
			{Key: "order_id", Value: []byte(event.OrderID)},
		},
	}

	err = writer.WriteMessages(ctx, message)
	if err != nil {
		p.logger.ErrorWithErr(err, fmt.Sprintf("Failed to publish event to topic %s", topic))
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Infof("Published event %s for payment %s to topic %s", event.EventType, event.PaymentID, topic)

	return nil
}

// Close closes all Kafka writers
func (p *EventPublisher) Close() error {
	var lastErr error
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.ErrorWithErr(err, fmt.Sprintf("Failed to close writer for topic %s", topic))
			lastErr = err
		}
	}
	return lastErr
}

// getTopicForEventType maps event type to Kafka topic
func (p *EventPublisher) getTopicForEventType(eventType string) string {
	topicMap := map[string]string{
		EventTypePaymentAuthorized: TopicPaymentAuthorized,
		EventTypePaymentCaptured:   TopicPaymentCaptured,
		EventTypePaymentFailed:     TopicPaymentFailed,
		EventTypePaymentCancelled:  TopicPaymentCancelled,
		EventTypePaymentRefunded:   TopicPaymentRefunded,
	}

	if topic, exists := topicMap[eventType]; exists {
		return topic
	}

	// Default to a generic payment events topic
	return "payment.events"
}

// hashString creates a hash value from a string (for partitioning)
func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
