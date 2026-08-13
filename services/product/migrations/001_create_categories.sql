-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    parent_id VARCHAR(36),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL
);

-- Create index on parent_id for faster subcategory queries
CREATE INDEX idx_categories_parent_id ON categories(parent_id);

-- Create index on name for faster lookups
CREATE INDEX idx_categories_name ON categories(name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_categories_name;
DROP INDEX IF EXISTS idx_categories_parent_id;
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd
