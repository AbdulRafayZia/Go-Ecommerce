package domain

import "errors"

// Domain errors for the order service
var (
	// ErrOrderNotFound is returned when an order is not found
	ErrOrderNotFound = errors.New("order not found")

	// ErrInvalidOrderStatus is returned when an invalid order status is provided
	ErrInvalidOrderStatus = errors.New("invalid order status")

	// ErrInvalidStateTransition is returned when an invalid state transition is attempted
	ErrInvalidStateTransition = errors.New("invalid state transition")

	// ErrEmptyOrder is returned when trying to create an order with no items
	ErrEmptyOrder = errors.New("order must have at least one item")

	// ErrInvalidUserID is returned when an invalid user ID is provided
	ErrInvalidUserID = errors.New("user ID cannot be empty")

	// ErrInvalidProductID is returned when an invalid product ID is provided
	ErrInvalidProductID = errors.New("product ID cannot be empty")

	// ErrInvalidQuantity is returned when an invalid quantity is provided
	ErrInvalidQuantity = errors.New("quantity must be greater than zero")

	// ErrInvalidPrice is returned when an invalid price is provided
	ErrInvalidPrice = errors.New("price must be greater than or equal to zero")

	// ErrOrderAlreadyPaid is returned when trying to pay an already paid order
	ErrOrderAlreadyPaid = errors.New("order is already paid")

	// ErrOrderNotPaid is returned when trying to ship an unpaid order
	ErrOrderNotPaid = errors.New("order has not been paid")

	// ErrOrderAlreadyCancelled is returned when trying to modify a cancelled order
	ErrOrderAlreadyCancelled = errors.New("order has been cancelled")

	// ErrOrderAlreadyDelivered is returned when trying to modify a delivered order
	ErrOrderAlreadyDelivered = errors.New("order has already been delivered")

	// ErrCannotCancelOrder is returned when trying to cancel an order that cannot be cancelled
	ErrCannotCancelOrder = errors.New("order cannot be cancelled in current state")

	// ErrInvalidIdempotencyKey is returned when an invalid idempotency key is provided
	ErrInvalidIdempotencyKey = errors.New("idempotency key cannot be empty")
)
