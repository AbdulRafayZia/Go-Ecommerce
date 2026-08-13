-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    price_cents INTEGER NOT NULL CHECK (price_cents >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    category_id VARCHAR(36),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    image_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);

-- Create index on category_id for faster filtering
CREATE INDEX idx_products_category_id ON products(category_id);

-- Create index on active for filtering active products
CREATE INDEX idx_products_active ON products(active);

-- Create index on name for search queries
CREATE INDEX idx_products_name ON products(name);

-- Create index on price_cents for price range filtering
CREATE INDEX idx_products_price_cents ON products(price_cents);

-- Create index on created_at for sorting
CREATE INDEX idx_products_created_at ON products(created_at DESC);

-- Create composite index for common query patterns
CREATE INDEX idx_products_active_category ON products(active, category_id);

-- Create GIN index for full-text search on name and description (PostgreSQL specific)
CREATE INDEX idx_products_search ON products USING GIN (
    to_tsvector('english', COALESCE(name, '') || ' ' || COALESCE(description, ''))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_products_search;
DROP INDEX IF EXISTS idx_products_active_category;
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_products_price_cents;
DROP INDEX IF EXISTS idx_products_name;
DROP INDEX IF EXISTS idx_products_active;
DROP INDEX IF EXISTS idx_products_category_id;
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
