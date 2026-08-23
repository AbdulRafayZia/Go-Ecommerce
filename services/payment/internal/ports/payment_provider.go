package ports

import (
	"context"
)

// PaymentProviderResult represents the result of a payment operation from the provider
type PaymentProviderResult struct {
	ProviderPaymentID string
	Status            string // "authorized", "captured", "failed"
	ErrorMessage      string
}

// PaymentProvider defines the interface for payment provider integration
// This abstracts payment providers like Stripe, PayPal, etc.
type PaymentProvider interface {
	// Authorize authorizes a payment (reserves funds but doesn't charge)
	Authorize(ctx context.Context, amount float64, currency string, metadata map[string]interface{}) (*PaymentProviderResult, error)

	// Capture captures an authorized payment (actually charges the customer)
	Capture(ctx context.Context, providerPaymentID string) (*PaymentProviderResult, error)

	// Cancel cancels an authorized payment (releases reserved funds)
	Cancel(ctx context.Context, providerPaymentID string) error

	// Refund refunds a captured payment
	Refund(ctx context.Context, providerPaymentID string, amount float64, reason string) (*PaymentProviderResult, error)

	// GetPaymentStatus retrieves the current status of a payment from the provider
	GetPaymentStatus(ctx context.Context, providerPaymentID string) (*PaymentProviderResult, error)
}
