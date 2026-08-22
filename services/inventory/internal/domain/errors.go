package domain

import "errors"

// Domain errors for the inventory service
var (
	// ErrStockNotFound is returned when stock is not found for a product
	ErrStockNotFound = errors.New("stock not found")

	// ErrReservationNotFound is returned when a reservation is not found
	ErrReservationNotFound = errors.New("reservation not found")

	// ErrInsufficientStock is returned when there's not enough stock to reserve
	ErrInsufficientStock = errors.New("insufficient stock available")

	// ErrInvalidQuantity is returned when an invalid quantity is provided
	ErrInvalidQuantity = errors.New("quantity must be greater than zero")

	// ErrInvalidProductID is returned when an invalid product ID is provided
	ErrInvalidProductID = errors.New("product ID cannot be empty")

	// ErrInvalidOrderID is returned when an invalid order ID is provided
	ErrInvalidOrderID = errors.New("order ID cannot be empty")

	// ErrReservationAlreadyFulfilled is returned when trying to fulfill an already fulfilled reservation
	ErrReservationAlreadyFulfilled = errors.New("reservation already fulfilled")

	// ErrReservationAlreadyCancelled is returned when trying to cancel an already cancelled reservation
	ErrReservationAlreadyCancelled = errors.New("reservation already cancelled")

	// ErrNegativeStock is returned when stock would become negative
	ErrNegativeStock = errors.New("stock cannot be negative")
)
