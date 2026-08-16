package ports

import (
	"context"
	"time"

	"gocommerce/services/cart/internal/domain"
)

// CartRepository defines the interface for cart storage operations
// This is a port (interface) that will be implemented by adapters (e.g., Redis)
type CartRepository interface {
	// Save stores a cart with TTL
	Save(ctx context.Context, cart *domain.Cart, ttl time.Duration) error

	// Get retrieves a cart by user ID
	Get(ctx context.Context, userID string) (*domain.Cart, error)

	// Delete removes a cart
	Delete(ctx context.Context, userID string) error

	// Exists checks if a cart exists for a user
	Exists(ctx context.Context, userID string) (bool, error)
}
