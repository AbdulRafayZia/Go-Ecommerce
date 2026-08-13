package ports

import (
	"context"

	"gocommerce/services/product/internal/domain"
)

// CategoryRepository defines the interface for category storage operations
type CategoryRepository interface {
	// Create stores a new category
	Create(ctx context.Context, category *domain.Category) error

	// GetByID retrieves a category by its ID
	GetByID(ctx context.Context, id string) (*domain.Category, error)

	// Update updates an existing category
	Update(ctx context.Context, category *domain.Category) error

	// Delete deletes a category
	Delete(ctx context.Context, id string) error

	// List retrieves all categories
	List(ctx context.Context) ([]*domain.Category, error)

	// ExistsByName checks if a category with the given name exists
	ExistsByName(ctx context.Context, name string) (bool, error)

	// GetByParentID retrieves all categories with the given parent ID
	GetByParentID(ctx context.Context, parentID string) ([]*domain.Category, error)
}
