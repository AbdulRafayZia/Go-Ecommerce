package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"gocommerce/services/payment/internal/adapters/kafka"
	"gocommerce/services/payment/internal/domain"
	"gocommerce/services/payment/internal/ports"
)

// PaymentService handles payment operations
type PaymentService struct {
	paymentRepo     ports.PaymentRepository
	paymentProvider ports.PaymentProvider
	eventPublisher  *kafka.EventPublisher
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	paymentRepo ports.PaymentRepository,
	paymentProvider ports.PaymentProvider,
	eventPublisher *kafka.EventPublisher,
) *PaymentService {
	return &PaymentService{
		paymentRepo:     paymentRepo,
		paymentProvider: paymentProvider,
		eventPublisher:  eventPublisher,
	}
}

// CreatePayment creates a new payment (idempotent)
func (s *PaymentService) CreatePayment(ctx context.Context, orderID string, amount float64, currency, idempotencyKey string, method domain.PaymentMethod) (*domain.Payment, error) {
	// Check if payment already exists (idempotency)
	existingPayment, err := s.paymentRepo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err == nil {
		// Payment already exists, return it
		return existingPayment, nil
	}
	if err != domain.ErrPaymentNotFound {
		return nil, fmt.Errorf("failed to check existing payment: %w", err)
	}

	// Create new payment
	payment, err := domain.NewPayment(orderID, amount, currency, idempotencyKey, method)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	payment.ID = uuid.New().String()

	// Authorize payment with provider
	result, err := s.paymentProvider.Authorize(ctx, amount, currency, map[string]interface{}{
		"order_id":    orderID,
		"payment_id":  payment.ID,
		"idempotency": idempotencyKey,
	})
	if err != nil {
		return nil, fmt.Errorf("payment provider error: %w", err)
	}

	// Handle provider response
	if result.Status == "failed" {
		payment.Fail(result.ErrorMessage)

		// Save failed payment
		if err := s.paymentRepo.Create(ctx, payment); err != nil {
			return nil, fmt.Errorf("failed to save payment: %w", err)
		}

		// Publish failure event
		event := kafka.NewPaymentEvent(
			kafka.EventTypePaymentFailed,
			payment.ID,
			payment.OrderID,
			payment.Amount,
			payment.Currency,
			string(payment.Status),
		)
		event.Metadata["failure_reason"] = result.ErrorMessage
		s.eventPublisher.PublishPaymentEvent(ctx, event)

		return payment, nil
	}

	// Authorize payment in domain
	if err := payment.Authorize(result.ProviderPaymentID); err != nil {
		return nil, fmt.Errorf("failed to authorize payment: %w", err)
	}

	// Save payment
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Publish authorized event
	event := kafka.NewPaymentEvent(
		kafka.EventTypePaymentAuthorized,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Currency,
		string(payment.Status),
	)
	event.Metadata["provider_payment_id"] = result.ProviderPaymentID
	s.eventPublisher.PublishPaymentEvent(ctx, event)

	return payment, nil
}

// GetPayment retrieves a payment by ID
func (s *PaymentService) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	return s.paymentRepo.GetByID(ctx, id)
}

// ListPayments lists payments with filtering
func (s *PaymentService) ListPayments(ctx context.Context, filter ports.PaymentFilter) ([]*domain.Payment, int, error) {
	return s.paymentRepo.List(ctx, filter)
}

// CapturePayment captures an authorized payment
func (s *PaymentService) CapturePayment(ctx context.Context, paymentID string) (*domain.Payment, error) {
	// Get payment
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// Check if payment can be captured
	if !payment.CanBeCaptured() {
		return nil, domain.ErrPaymentNotAuthorized
	}

	// Capture with provider
	result, err := s.paymentProvider.Capture(ctx, payment.ProviderPaymentID)
	if err != nil {
		return nil, fmt.Errorf("payment provider error: %w", err)
	}

	// Handle provider response
	if result.Status == "failed" {
		payment.Fail(result.ErrorMessage)

		// Update payment
		if err := s.paymentRepo.Update(ctx, payment); err != nil {
			return nil, fmt.Errorf("failed to update payment: %w", err)
		}

		// Publish failure event
		event := kafka.NewPaymentEvent(
			kafka.EventTypePaymentFailed,
			payment.ID,
			payment.OrderID,
			payment.Amount,
			payment.Currency,
			string(payment.Status),
		)
		event.Metadata["failure_reason"] = result.ErrorMessage
		s.eventPublisher.PublishPaymentEvent(ctx, event)

		return payment, nil
	}

	// Capture payment in domain
	if err := payment.Capture(); err != nil {
		return nil, fmt.Errorf("failed to capture payment: %w", err)
	}

	// Update payment
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Publish captured event
	event := kafka.NewPaymentEvent(
		kafka.EventTypePaymentCaptured,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Currency,
		string(payment.Status),
	)
	s.eventPublisher.PublishPaymentEvent(ctx, event)

	return payment, nil
}

// CancelPayment cancels a pending or authorized payment
func (s *PaymentService) CancelPayment(ctx context.Context, paymentID string) (*domain.Payment, error) {
	// Get payment
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// Check if payment can be cancelled
	if !payment.CanBeCancelled() {
		return nil, domain.ErrInvalidStateTransition
	}

	// Cancel with provider (if authorized)
	if payment.IsAuthorized() {
		if err := s.paymentProvider.Cancel(ctx, payment.ProviderPaymentID); err != nil {
			return nil, fmt.Errorf("payment provider error: %w", err)
		}
	}

	// Cancel payment in domain
	if err := payment.Cancel(); err != nil {
		return nil, fmt.Errorf("failed to cancel payment: %w", err)
	}

	// Update payment
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Publish cancelled event
	event := kafka.NewPaymentEvent(
		kafka.EventTypePaymentCancelled,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Currency,
		string(payment.Status),
	)
	s.eventPublisher.PublishPaymentEvent(ctx, event)

	return payment, nil
}

// RefundPayment refunds a captured payment
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID, reason string) (*domain.Payment, error) {
	// Get payment
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// Check if payment can be refunded
	if !payment.CanBeRefunded() {
		return nil, domain.ErrPaymentNotCaptured
	}

	// Refund with provider
	result, err := s.paymentProvider.Refund(ctx, payment.ProviderPaymentID, payment.Amount, reason)
	if err != nil {
		return nil, fmt.Errorf("payment provider error: %w", err)
	}

	// Handle provider response
	if result.Status == "failed" {
		return nil, fmt.Errorf("refund failed: %s", result.ErrorMessage)
	}

	// Refund payment in domain
	if err := payment.Refund(reason); err != nil {
		return nil, fmt.Errorf("failed to refund payment: %w", err)
	}

	// Update payment
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Publish refunded event
	event := kafka.NewPaymentEvent(
		kafka.EventTypePaymentRefunded,
		payment.ID,
		payment.OrderID,
		payment.Amount,
		payment.Currency,
		string(payment.Status),
	)
	event.Metadata["refund_reason"] = reason
	s.eventPublisher.PublishPaymentEvent(ctx, event)

	return payment, nil
}

// HandleWebhook handles webhook callbacks from payment provider
func (s *PaymentService) HandleWebhook(ctx context.Context, eventType, providerPaymentID string, data map[string]interface{}) error {
	// Get payment by provider payment ID
	payment, err := s.paymentRepo.GetByProviderPaymentID(ctx, providerPaymentID)
	if err != nil {
		return fmt.Errorf("payment not found for provider ID %s: %w", providerPaymentID, err)
	}

	// Handle different event types
	switch eventType {
	case "payment.captured":
		if !payment.IsCaptured() {
			payment.Capture()
			if err := s.paymentRepo.Update(ctx, payment); err != nil {
				return fmt.Errorf("failed to update payment: %w", err)
			}

			// Publish event
			event := kafka.NewPaymentEvent(
				kafka.EventTypePaymentCaptured,
				payment.ID,
				payment.OrderID,
				payment.Amount,
				payment.Currency,
				string(payment.Status),
			)
			s.eventPublisher.PublishPaymentEvent(ctx, event)
		}

	case "payment.failed":
		if !payment.IsFailed() {
			reason := "Unknown error"
			if msg, ok := data["error_message"].(string); ok {
				reason = msg
			}

			payment.Fail(reason)
			if err := s.paymentRepo.Update(ctx, payment); err != nil {
				return fmt.Errorf("failed to update payment: %w", err)
			}

			// Publish event
			event := kafka.NewPaymentEvent(
				kafka.EventTypePaymentFailed,
				payment.ID,
				payment.OrderID,
				payment.Amount,
				payment.Currency,
				string(payment.Status),
			)
			event.Metadata["failure_reason"] = reason
			s.eventPublisher.PublishPaymentEvent(ctx, event)
		}

	case "payment.refunded":
		if !payment.IsRefunded() {
			reason := ""
			if r, ok := data["reason"].(string); ok {
				reason = r
			}

			payment.Refund(reason)
			if err := s.paymentRepo.Update(ctx, payment); err != nil {
				return fmt.Errorf("failed to update payment: %w", err)
			}

			// Publish event
			event := kafka.NewPaymentEvent(
				kafka.EventTypePaymentRefunded,
				payment.ID,
				payment.OrderID,
				payment.Amount,
				payment.Currency,
				string(payment.Status),
			)
			s.eventPublisher.PublishPaymentEvent(ctx, event)
		}

	default:
		// Unknown event type, just log it
		return fmt.Errorf("unknown webhook event type: %s", eventType)
	}

	return nil
}
