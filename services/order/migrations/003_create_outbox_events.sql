-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS outbox_events (
    id BIGSERIAL PRIMARY KEY,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP,
    published BOOLEAN NOT NULL DEFAULT FALSE
);

-- Create index on published for faster unpublished event queries
CREATE INDEX idx_outbox_events_published ON outbox_events(published, created_at) WHERE published = FALSE;

-- Create index on aggregate for event sourcing queries
CREATE INDEX idx_outbox_events_aggregate ON outbox_events(aggregate_type, aggregate_id);

-- Create index on created_at for ordering
CREATE INDEX idx_outbox_events_created_at ON outbox_events(created_at ASC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_outbox_events_created_at;
DROP INDEX IF EXISTS idx_outbox_events_aggregate;
DROP INDEX IF EXISTS idx_outbox_events_published;
DROP TABLE IF EXISTS outbox_events;
-- +goose StatementEnd
