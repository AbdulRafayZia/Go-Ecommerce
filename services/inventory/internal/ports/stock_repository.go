package ports

import (
	"context"

	"gocommerce/services/inventory/internal/domain"
)

// StockRepository defines the interface for stock persistence
type StockRepository interface {
	// GetByProductID retrieves stock information for a product
	GetByProductID(ctx context.Context, productID string) (*domain.Stock, error)

	// Create creates a new stock record
	Create(ctx context.Context, stock *domain.Stock) error

	// Update updates an existing stock record
	Update(ctx context.Context, stock *domain.Stock) error

	// CreateOrUpdate creates or updates a stock record (upsert)
	CreateOrUpdate(ctx context.Context, stock *domain.Stock) error

	// ReserveStock reserves stock for an order (atomic operation)
	// This method handles the stock reservation within a transaction
	ReserveStock(ctx context.Context, productID string, quantity int) error

	// ReleaseReservation releases reserved stock back to available
	ReleaseReservation(ctx context.Context, productID string, quantity int) error

	// FulfillReservation deducts reserved stock (when order is paid)
	FulfillReservation(ctx context.Context, productID string, quantity int) error

	// ListLowStock returns products with available stock below threshold
	ListLowStock(ctx context.Context, threshold int) ([]*domain.Stock, error)
}
