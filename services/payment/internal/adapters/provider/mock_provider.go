package provider

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"

	"gocommerce/pkg/logger"
	"gocommerce/services/payment/internal/ports"
)

// MockPaymentProvider is a mock implementation of PaymentProvider for testing
// In production, this would be replaced with real payment provider implementations (Stripe, PayPal, etc.)
type MockPaymentProvider struct {
	logger           *logger.Logger
	failureRate      float64 // Probability of payment failure (0.0 to 1.0)
	authorizeEnabled bool    // If false, directly captures payments
}

// NewMockPaymentProvider creates a new mock payment provider
func NewMockPaymentProvider(log *logger.Logger, failureRate float64) ports.PaymentProvider {
	return &MockPaymentProvider{
		logger:           log,
		failureRate:      failureRate,
		authorizeEnabled: true, // Default to two-step auth + capture flow
	}
}

// Authorize simulates authorizing a payment (reserving funds)
func (m *MockPaymentProvider) Authorize(ctx context.Context, amount float64, currency string, metadata map[string]interface{}) (*ports.PaymentProviderResult, error) {
	m.logger.Infof("Mock Provider: Authorizing payment for amount %.2f %s", amount, currency)

	// Simulate processing delay
	// In real implementation, this would call external API

	// Generate mock provider payment ID
	providerPaymentID := fmt.Sprintf("mock_auth_%s", uuid.New().String()[:8])

	// Simulate random failures based on failure rate
	if rand.Float64() < m.failureRate {
		m.logger.Warnf("Mock Provider: Payment authorization failed for %s", providerPaymentID)
		return &ports.PaymentProviderResult{
			ProviderPaymentID: providerPaymentID,
			Status:            "failed",
			ErrorMessage:      "Insufficient funds",
		}, nil
	}

	m.logger.Infof("Mock Provider: Payment authorized successfully: %s", providerPaymentID)

	return &ports.PaymentProviderResult{
		ProviderPaymentID: providerPaymentID,
		Status:            "authorized",
		ErrorMessage:      "",
	}, nil
}

// Capture simulates capturing an authorized payment (actually charging)
func (m *MockPaymentProvider) Capture(ctx context.Context, providerPaymentID string) (*ports.PaymentProviderResult, error) {
	m.logger.Infof("Mock Provider: Capturing payment %s", providerPaymentID)

	// In real implementation, this would call provider API to capture the authorized payment
	// For Stripe: stripe.PaymentIntents.Capture(paymentIntentID)

	// Simulate small failure chance even for capture
	if rand.Float64() < (m.failureRate * 0.1) { // 10% of failure rate
		m.logger.Warnf("Mock Provider: Payment capture failed for %s", providerPaymentID)
		return &ports.PaymentProviderResult{
			ProviderPaymentID: providerPaymentID,
			Status:            "failed",
			ErrorMessage:      "Capture declined by bank",
		}, nil
	}

	m.logger.Infof("Mock Provider: Payment captured successfully: %s", providerPaymentID)

	return &ports.PaymentProviderResult{
		ProviderPaymentID: providerPaymentID,
		Status:            "captured",
		ErrorMessage:      "",
	}, nil
}

// Cancel simulates canceling an authorized payment (releasing funds)
func (m *MockPaymentProvider) Cancel(ctx context.Context, providerPaymentID string) error {
	m.logger.Infof("Mock Provider: Canceling payment %s", providerPaymentID)

	// In real implementation, this would call provider API
	// For Stripe: stripe.PaymentIntents.Cancel(paymentIntentID)

	m.logger.Infof("Mock Provider: Payment cancelled successfully: %s", providerPaymentID)

	return nil
}

// Refund simulates refunding a captured payment
func (m *MockPaymentProvider) Refund(ctx context.Context, providerPaymentID string, amount float64, reason string) (*ports.PaymentProviderResult, error) {
	m.logger.Infof("Mock Provider: Refunding payment %s for amount %.2f (reason: %s)", providerPaymentID, amount, reason)

	// In real implementation, this would call provider API
	// For Stripe: stripe.Refunds.Create(&stripe.RefundParams{PaymentIntent: paymentIntentID})

	// Simulate small failure chance
	if rand.Float64() < (m.failureRate * 0.05) { // 5% of failure rate
		m.logger.Warnf("Mock Provider: Refund failed for %s", providerPaymentID)
		return &ports.PaymentProviderResult{
			ProviderPaymentID: providerPaymentID,
			Status:            "failed",
			ErrorMessage:      "Refund processing error",
		}, nil
	}

	m.logger.Infof("Mock Provider: Payment refunded successfully: %s", providerPaymentID)

	return &ports.PaymentProviderResult{
		ProviderPaymentID: providerPaymentID,
		Status:            "refunded",
		ErrorMessage:      "",
	}, nil
}

// GetPaymentStatus retrieves the current status from the provider
func (m *MockPaymentProvider) GetPaymentStatus(ctx context.Context, providerPaymentID string) (*ports.PaymentProviderResult, error) {
	m.logger.Infof("Mock Provider: Getting payment status for %s", providerPaymentID)

	// In real implementation, this would call provider API
	// For Stripe: stripe.PaymentIntents.Get(paymentIntentID)

	// Mock: Just return a successful status
	// In reality, this would query the actual provider
	return &ports.PaymentProviderResult{
		ProviderPaymentID: providerPaymentID,
		Status:            "authorized",
		ErrorMessage:      "",
	}, nil
}
