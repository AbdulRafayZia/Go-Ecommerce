package ports

import (
	"context"

	"gocommerce/services/order/internal/domain"
)

// OrderRepository defines the interface for order storage operations
type OrderRepository interface {
	// Create stores a new order and its items in a transaction
	// Also stores any domain events in the outbox table
	Create(ctx context.Context, order *domain.Order) error

	// GetByID retrieves an order by its ID
	GetByID(ctx context.Context, id string) (*domain.Order, error)

	// GetByIdempotencyKey retrieves an order by idempotency key
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Order, error)

	// Update updates an existing order and stores any new domain events
	Update(ctx context.Context, order *domain.Order) error

	// List retrieves orders with pagination and filtering
	List(ctx context.Context, filters ListOrderFilters) ([]*domain.Order, int64, error)

	// ExistsByIdempotencyKey checks if an order with the given idempotency key exists
	ExistsByIdempotencyKey(ctx context.Context, key string) (bool, error)
}

// ListOrderFilters defines filtering options for listing orders
type ListOrderFilters struct {
	Page     int
	PageSize int
	UserID   *string
	Status   *domain.OrderStatus
}
