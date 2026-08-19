package ports

import (
	"context"
	"time"
)

// OutboxEvent represents an event stored in the outbox table
type OutboxEvent struct {
	ID            int64
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte // JSON payload
	CreatedAt     time.Time
	PublishedAt   *time.Time
	Published     bool
}

// OutboxRepository defines the interface for outbox event storage
type OutboxRepository interface {
	// GetUnpublishedEvents retrieves unpublished events up to the given limit
	GetUnpublishedEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)

	// MarkAsPublished marks an event as published
	MarkAsPublished(ctx context.Context, eventID int64) error

	// MarkMultipleAsPublished marks multiple events as published in a batch
	MarkMultipleAsPublished(ctx context.Context, eventIDs []int64) error
}
