package domain

import (
	"strings"
	"time"
)

// Stock represents the inventory stock for a product
type Stock struct {
	ProductID string
	Available int // Available for purchase (not reserved)
	Reserved  int // Reserved for orders but not yet deducted
	UpdatedAt time.Time
}

// NewStock creates a new stock record with validation
func NewStock(productID string, initialQuantity int) (*Stock, error) {
	if strings.TrimSpace(productID) == "" {
		return nil, ErrInvalidProductID
	}

	if initialQuantity < 0 {
		return nil, ErrNegativeStock
	}

	return &Stock{
		ProductID: productID,
		Available: initialQuantity,
		Reserved:  0,
		UpdatedAt: time.Now(),
	}, nil
}

// TotalStock returns the total stock (available + reserved)
func (s *Stock) TotalStock() int {
	return s.Available + s.Reserved
}

// CanReserve checks if the requested quantity can be reserved
func (s *Stock) CanReserve(quantity int) bool {
	return s.Available >= quantity
}

// Reserve reserves stock for an order
func (s *Stock) Reserve(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if !s.CanReserve(quantity) {
		return ErrInsufficientStock
	}

	s.Available -= quantity
	s.Reserved += quantity
	s.UpdatedAt = time.Now()

	return nil
}

// ReleaseReservation releases reserved stock back to available
func (s *Stock) ReleaseReservation(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if s.Reserved < quantity {
		return ErrInvalidQuantity // Trying to release more than reserved
	}

	s.Reserved -= quantity
	s.Available += quantity
	s.UpdatedAt = time.Now()

	return nil
}

// FulfillReservation deducts reserved stock (order was paid/shipped)
func (s *Stock) FulfillReservation(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if s.Reserved < quantity {
		return ErrInvalidQuantity // Trying to fulfill more than reserved
	}

	s.Reserved -= quantity
	s.UpdatedAt = time.Now()

	return nil
}

// AddStock adds new stock (restocking)
func (s *Stock) AddStock(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	s.Available += quantity
	s.UpdatedAt = time.Now()

	return nil
}

// DeductStock directly deducts from available stock (for manual adjustments)
func (s *Stock) DeductStock(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	if s.Available < quantity {
		return ErrInsufficientStock
	}

	s.Available -= quantity
	s.UpdatedAt = time.Now()

	return nil
}

// SetStock sets the available stock to a specific value (for inventory adjustments)
func (s *Stock) SetStock(quantity int) error {
	if quantity < 0 {
		return ErrNegativeStock
	}

	s.Available = quantity
	s.UpdatedAt = time.Now()

	return nil
}
