package ports

import (
	"context"
	"time"

	"gocommerce/services/inventory/internal/domain"
)

// ReservationRepository defines the interface for reservation persistence
type ReservationRepository interface {
	// Create creates a new reservation
	Create(ctx context.Context, reservation *domain.Reservation) error

	// GetByID retrieves a reservation by its ID
	GetByID(ctx context.Context, id string) (*domain.Reservation, error)

	// GetByOrderID retrieves all reservations for an order
	GetByOrderID(ctx context.Context, orderID string) ([]*domain.Reservation, error)

	// GetByOrderAndProduct retrieves a specific reservation for an order and product
	GetByOrderAndProduct(ctx context.Context, orderID, productID string) (*domain.Reservation, error)

	// Update updates an existing reservation
	Update(ctx context.Context, reservation *domain.Reservation) error

	// ListPendingReservations lists all pending reservations
	ListPendingReservations(ctx context.Context, limit int) ([]*domain.Reservation, error)

	// ListExpiredReservations lists reservations that have been pending longer than the duration
	ListExpiredReservations(ctx context.Context, olderThan time.Duration) ([]*domain.Reservation, error)

	// Delete deletes a reservation (cleanup after fulfillment/cancellation)
	Delete(ctx context.Context, id string) error
}
