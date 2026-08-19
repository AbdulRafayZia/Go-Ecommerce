package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gocommerce/services/order/internal/ports"
)

// OutboxRepository is the PostgreSQL implementation of ports.OutboxRepository
type OutboxRepository struct {
	db *sql.DB
}

// NewOutboxRepository creates a new PostgreSQL outbox repository
func NewOutboxRepository(db *sql.DB) ports.OutboxRepository {
	return &OutboxRepository{db: db}
}

// GetUnpublishedEvents retrieves unpublished events up to the given limit
func (r *OutboxRepository) GetUnpublishedEvents(ctx context.Context, limit int) ([]*ports.OutboxEvent, error) {
	query := `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, created_at, published_at, published
		FROM outbox_events
		WHERE published = FALSE
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query unpublished events: %w", err)
	}
	defer rows.Close()

	events := make([]*ports.OutboxEvent, 0)
	for rows.Next() {
		event := &ports.OutboxEvent{}
		var publishedAt sql.NullTime

		if err := rows.Scan(
			&event.ID,
			&event.AggregateType,
			&event.AggregateID,
			&event.EventType,
			&event.Payload,
			&event.CreatedAt,
			&publishedAt,
			&event.Published,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outbox event: %w", err)
		}

		if publishedAt.Valid {
			event.PublishedAt = &publishedAt.Time
		}

		events = append(events, event)
	}

	return events, nil
}

// MarkAsPublished marks an event as published
func (r *OutboxRepository) MarkAsPublished(ctx context.Context, eventID int64) error {
	query := `
		UPDATE outbox_events
		SET published = TRUE, published_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event as published: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("event %d not found", eventID)
	}

	return nil
}

// MarkMultipleAsPublished marks multiple events as published in a batch
func (r *OutboxRepository) MarkMultipleAsPublished(ctx context.Context, eventIDs []int64) error {
	if len(eventIDs) == 0 {
		return nil
	}

	// Start transaction for batch update
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		UPDATE outbox_events
		SET published = TRUE, published_at = $1
		WHERE id = $2
	`

	now := time.Now()
	for _, eventID := range eventIDs {
		_, err := tx.ExecContext(ctx, query, now, eventID)
		if err != nil {
			return fmt.Errorf("failed to mark event %d as published: %w", eventID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
