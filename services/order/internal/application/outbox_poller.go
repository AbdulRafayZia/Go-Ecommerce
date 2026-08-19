package application

import (
	"context"
	"fmt"
	"time"

	"gocommerce/pkg/logger"
	"gocommerce/services/order/internal/ports"
)

// OutboxPoller polls the outbox for unpublished events and publishes them
type OutboxPoller struct {
	outboxRepo ports.OutboxRepository
	publisher  ports.EventPublisher
	logger     *logger.Logger
	pollInterval time.Duration
	batchSize    int
}

// NewOutboxPoller creates a new outbox poller
func NewOutboxPoller(
	outboxRepo ports.OutboxRepository,
	publisher ports.EventPublisher,
	logger *logger.Logger,
	pollInterval time.Duration,
	batchSize int,
) *OutboxPoller {
	return &OutboxPoller{
		outboxRepo:   outboxRepo,
		publisher:    publisher,
		logger:       logger,
		pollInterval: pollInterval,
		batchSize:    batchSize,
	}
}

// Start starts the outbox poller (runs in background)
func (p *OutboxPoller) Start(ctx context.Context) {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	p.logger.Info("Outbox poller started")

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Outbox poller stopped")
			return
		case <-ticker.C:
			if err := p.pollAndPublish(ctx); err != nil {
				p.logger.ErrorWithErr(err, "Failed to poll and publish events")
			}
		}
	}
}

// pollAndPublish retrieves unpublished events and publishes them to Kafka
func (p *OutboxPoller) pollAndPublish(ctx context.Context) error {
	// Get unpublished events
	events, err := p.outboxRepo.GetUnpublishedEvents(ctx, p.batchSize)
	if err != nil {
		return fmt.Errorf("failed to get unpublished events: %w", err)
	}

	if len(events) == 0 {
		return nil // No events to publish
	}

	p.logger.Infof("Processing %d unpublished events", len(events))

	// Publish events and track published IDs
	publishedIDs := make([]int64, 0, len(events))

	for _, event := range events {
		// Determine topic based on event type
		topic := p.getTopicForEventType(event.EventType)

		// Publish to Kafka
		err := p.publisher.Publish(ctx, topic, event.AggregateID, event.Payload)
		if err != nil {
			p.logger.ErrorWithErr(err, fmt.Sprintf("Failed to publish event %d to Kafka", event.ID))
			// Continue with other events instead of failing the entire batch
			continue
		}

		publishedIDs = append(publishedIDs, event.ID)

		p.logger.Infof("Published event %d: type=%s, topic=%s, aggregate=%s",
			event.ID, event.EventType, topic, event.AggregateID)
	}

	// Mark successfully published events
	if len(publishedIDs) > 0 {
		if err := p.outboxRepo.MarkMultipleAsPublished(ctx, publishedIDs); err != nil {
			p.logger.ErrorWithErr(err, "Failed to mark events as published")
			return err
		}

		p.logger.Infof("Marked %d events as published", len(publishedIDs))
	}

	return nil
}

// getTopicForEventType maps event types to Kafka topics
func (p *OutboxPoller) getTopicForEventType(eventType string) string {
	// Map event types to topics
	// In a real system, this might be more sophisticated
	switch eventType {
	case "order.created", "order.paid", "order.shipped", "order.delivered", "order.cancelled", "order.failed":
		return "orders"
	default:
		return "orders" // Default topic
	}
}
