package ports

import (
	"context"

	"gocommerce/services/product/internal/domain"
)

// ProductRepository defines the interface for product storage operations
// This is a port (interface) that will be implemented by adapters (e.g., PostgreSQL)
type ProductRepository interface {
	// Create stores a new product
	Create(ctx context.Context, product *domain.Product) error

	// GetByID retrieves a product by its ID
	GetByID(ctx context.Context, id string) (*domain.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, product *domain.Product) error

	// Delete soft deletes a product (sets active = false)
	Delete(ctx context.Context, id string) error

	// List retrieves products with pagination and filtering
	List(ctx context.Context, filters ListProductFilters) ([]*domain.Product, int64, error)

	// ExistsByName checks if a product with the given name exists
	ExistsByName(ctx context.Context, name string) (bool, error)
}

// ListProductFilters defines filtering options for listing products
type ListProductFilters struct {
	Page       int
	PageSize   int
	CategoryID *string
	Search     *string
	MinPrice   *int
	MaxPrice   *int
	ActiveOnly bool
}
