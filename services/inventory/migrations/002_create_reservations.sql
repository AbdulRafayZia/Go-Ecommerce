-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS reservations (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    product_id VARCHAR(36) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (product_id) REFERENCES stock(product_id) ON DELETE RESTRICT
);

-- Create index on order_id for faster order reservation queries
CREATE INDEX idx_reservations_order_id ON reservations(order_id);

-- Create index on product_id for faster product reservation queries
CREATE INDEX idx_reservations_product_id ON reservations(product_id);

-- Create index on status for filtering pending reservations
CREATE INDEX idx_reservations_status ON reservations(status);

-- Create composite index for pending reservations cleanup
CREATE INDEX idx_reservations_pending ON reservations(status, created_at) WHERE status = 'pending';

-- Create unique index to prevent duplicate reservations for the same order and product
CREATE UNIQUE INDEX idx_reservations_order_product ON reservations(order_id, product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reservations_order_product;
DROP INDEX IF EXISTS idx_reservations_pending;
DROP INDEX IF EXISTS idx_reservations_status;
DROP INDEX IF EXISTS idx_reservations_product_id;
DROP INDEX IF EXISTS idx_reservations_order_id;
DROP TABLE IF EXISTS reservations;
-- +goose StatementEnd
