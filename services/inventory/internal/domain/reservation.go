package domain

import (
	"strings"
	"time"
)

// Reservation represents a stock reservation for an order
type Reservation struct {
	ID        string
	OrderID   string
	ProductID string
	Quantity  int
	Status    ReservationStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewReservation creates a new stock reservation
func NewReservation(orderID, productID string, quantity int) (*Reservation, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, ErrInvalidOrderID
	}

	if strings.TrimSpace(productID) == "" {
		return nil, ErrInvalidProductID
	}

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	now := time.Now()
	return &Reservation{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Status:    ReservationStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Fulfill marks the reservation as fulfilled (stock was deducted)
func (r *Reservation) Fulfill() error {
	if r.Status == ReservationStatusFulfilled {
		return ErrReservationAlreadyFulfilled
	}

	if r.Status.IsFinal() {
		return ErrReservationAlreadyCancelled
	}

	r.Status = ReservationStatusFulfilled
	r.UpdatedAt = time.Now()

	return nil
}

// Cancel marks the reservation as cancelled (stock was released)
func (r *Reservation) Cancel() error {
	if r.Status == ReservationStatusCancelled {
		return ErrReservationAlreadyCancelled
	}

	if r.Status.IsFinal() {
		return ErrReservationAlreadyFulfilled
	}

	r.Status = ReservationStatusCancelled
	r.UpdatedAt = time.Now()

	return nil
}

// Expire marks the reservation as expired (stock was released)
func (r *Reservation) Expire() error {
	if r.Status.IsFinal() {
		return nil // Already in final state, ignore
	}

	r.Status = ReservationStatusExpired
	r.UpdatedAt = time.Now()

	return nil
}

// IsPending checks if the reservation is still pending
func (r *Reservation) IsPending() bool {
	return r.Status == ReservationStatusPending
}

// IsFinal checks if the reservation is in a final state
func (r *Reservation) IsFinal() bool {
	return r.Status.IsFinal()
}
