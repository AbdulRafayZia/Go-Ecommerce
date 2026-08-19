-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    status VARCHAR(50) NOT NULL,
    total_cents INTEGER NOT NULL CHECK (total_cents >= 0),
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    payment_id VARCHAR(36),
    tracking_number VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on user_id for faster user order queries
CREATE INDEX idx_orders_user_id ON orders(user_id);

-- Create index on status for filtering by status
CREATE INDEX idx_orders_status ON orders(status);

-- Create index on idempotency_key for duplicate prevention
CREATE INDEX idx_orders_idempotency_key ON orders(idempotency_key);

-- Create index on created_at for sorting
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);

-- Create composite index for common query patterns
CREATE INDEX idx_orders_user_status ON orders(user_id, status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_orders_user_status;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_idempotency_key;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
