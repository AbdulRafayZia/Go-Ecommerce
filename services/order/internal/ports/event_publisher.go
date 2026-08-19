package ports

import "context"

// EventPublisher defines the interface for publishing domain events
// This will be implemented by the Kafka adapter
type EventPublisher interface {
	// Publish publishes an event to the message broker
	Publish(ctx context.Context, topic string, key string, payload []byte) error

	// Close closes the publisher connection
	Close() error
}
