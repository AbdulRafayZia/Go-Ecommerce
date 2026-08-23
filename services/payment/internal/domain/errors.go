package domain

import "errors"

// Domain errors for the payment service
var (
	// ErrPaymentNotFound is returned when a payment is not found
	ErrPaymentNotFound = errors.New("payment not found")

	// ErrInvalidAmount is returned when an invalid amount is provided
	ErrInvalidAmount = errors.New("amount must be greater than zero")

	// ErrInvalidCurrency is returned when an invalid currency is provided
	ErrInvalidCurrency = errors.New("currency must be a valid 3-letter code")

	// ErrInvalidOrderID is returned when an invalid order ID is provided
	ErrInvalidOrderID = errors.New("order ID cannot be empty")

	// ErrPaymentAlreadyExists is returned when a payment already exists for an idempotency key
	ErrPaymentAlreadyExists = errors.New("payment already exists for this idempotency key")

	// ErrInvalidStateTransition is returned when an invalid state transition is attempted
	ErrInvalidStateTransition = errors.New("invalid payment state transition")

	// ErrPaymentNotAuthorized is returned when trying to capture a payment that hasn't been authorized
	ErrPaymentNotAuthorized = errors.New("payment is not authorized")

	// ErrPaymentNotCaptured is returned when trying to refund a payment that hasn't been captured
	ErrPaymentNotCaptured = errors.New("payment is not captured")

	// ErrPaymentAlreadyCaptured is returned when trying to capture an already captured payment
	ErrPaymentAlreadyCaptured = errors.New("payment already captured")

	// ErrPaymentAlreadyCancelled is returned when trying to cancel an already cancelled payment
	ErrPaymentAlreadyCancelled = errors.New("payment already cancelled")

	// ErrPaymentAlreadyRefunded is returned when trying to refund an already refunded payment
	ErrPaymentAlreadyRefunded = errors.New("payment already refunded")

	// ErrProviderError is returned when the payment provider returns an error
	ErrProviderError = errors.New("payment provider error")
)
