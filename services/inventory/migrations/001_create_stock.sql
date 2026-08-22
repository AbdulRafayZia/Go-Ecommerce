-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS stock (
    product_id VARCHAR(36) PRIMARY KEY,
    available INTEGER NOT NULL DEFAULT 0 CHECK (available >= 0),
    reserved INTEGER NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on updated_at for tracking changes
CREATE INDEX idx_stock_updated_at ON stock(updated_at DESC);

-- Create index for low stock queries
CREATE INDEX idx_stock_available ON stock(available) WHERE available < 10;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_stock_available;
DROP INDEX IF EXISTS idx_stock_updated_at;
DROP TABLE IF EXISTS stock;
-- +goose StatementEnd
