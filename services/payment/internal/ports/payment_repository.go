package ports

import (
	"context"

	"gocommerce/services/payment/internal/domain"
)

// PaymentRepository defines the interface for payment persistence
type PaymentRepository interface {
	// Create creates a new payment
	Create(ctx context.Context, payment *domain.Payment) error

	// GetByID retrieves a payment by its ID
	GetByID(ctx context.Context, id string) (*domain.Payment, error)

	// GetByIdempotencyKey retrieves a payment by its idempotency key
	GetByIdempotencyKey(ctx context.Context, key string) (*domain.Payment, error)

	// GetByOrderID retrieves the payment for an order
	GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error)

	// GetByProviderPaymentID retrieves a payment by provider payment ID
	GetByProviderPaymentID(ctx context.Context, providerID string) (*domain.Payment, error)

	// Update updates an existing payment
	Update(ctx context.Context, payment *domain.Payment) error

	// List retrieves payments with optional filtering
	List(ctx context.Context, filter PaymentFilter) ([]*domain.Payment, int, error)
}

// PaymentFilter defines filters for listing payments
type PaymentFilter struct {
	OrderID *string
	Status  *domain.PaymentStatus
	Limit   int
	Offset  int
}
