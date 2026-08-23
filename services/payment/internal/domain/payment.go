package domain

import (
	"strings"
	"time"
)

// PaymentMethod represents the payment method type
type PaymentMethod string

const (
	PaymentMethodCard         PaymentMethod = "card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodWallet       PaymentMethod = "wallet"
)

// Payment represents a payment aggregate root
type Payment struct {
	ID                 string
	OrderID            string
	Amount             float64
	Currency           string
	Status             PaymentStatus
	PaymentMethod      PaymentMethod
	ProviderPaymentID  string // ID from payment provider (e.g., Stripe charge ID)
	IdempotencyKey     string
	FailureReason      string
	Metadata           map[string]interface{}
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewPayment creates a new payment with validation
func NewPayment(orderID string, amount float64, currency, idempotencyKey string, method PaymentMethod) (*Payment, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, ErrInvalidOrderID
	}

	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if len(currency) != 3 {
		return nil, ErrInvalidCurrency
	}

	now := time.Now()

	return &Payment{
		OrderID:        orderID,
		Amount:         amount,
		Currency:       strings.ToUpper(currency),
		Status:         PaymentStatusPending,
		PaymentMethod:  method,
		IdempotencyKey: idempotencyKey,
		Metadata:       make(map[string]interface{}),
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

// Authorize marks the payment as authorized (funds reserved but not captured)
func (p *Payment) Authorize(providerPaymentID string) error {
	if !p.Status.CanTransitionTo(PaymentStatusAuthorized) {
		return ErrInvalidStateTransition
	}

	p.Status = PaymentStatusAuthorized
	p.ProviderPaymentID = providerPaymentID
	p.UpdatedAt = time.Now()

	return nil
}

// Capture marks the payment as captured (funds charged to customer)
func (p *Payment) Capture() error {
	if p.Status == PaymentStatusCaptured {
		return ErrPaymentAlreadyCaptured
	}

	if !p.Status.CanTransitionTo(PaymentStatusCaptured) {
		return ErrInvalidStateTransition
	}

	p.Status = PaymentStatusCaptured
	p.UpdatedAt = time.Now()

	return nil
}

// Fail marks the payment as failed
func (p *Payment) Fail(reason string) error {
	if !p.Status.CanTransitionTo(PaymentStatusFailed) {
		return ErrInvalidStateTransition
	}

	p.Status = PaymentStatusFailed
	p.FailureReason = reason
	p.UpdatedAt = time.Now()

	return nil
}

// Cancel marks the payment as cancelled
func (p *Payment) Cancel() error {
	if p.Status == PaymentStatusCancelled {
		return ErrPaymentAlreadyCancelled
	}

	if !p.Status.CanTransitionTo(PaymentStatusCancelled) {
		return ErrInvalidStateTransition
	}

	p.Status = PaymentStatusCancelled
	p.UpdatedAt = time.Now()

	return nil
}

// Refund marks the payment as refunded
func (p *Payment) Refund(reason string) error {
	if p.Status == PaymentStatusRefunded {
		return ErrPaymentAlreadyRefunded
	}

	if p.Status != PaymentStatusCaptured {
		return ErrPaymentNotCaptured
	}

	if !p.Status.CanTransitionTo(PaymentStatusRefunded) {
		return ErrInvalidStateTransition
	}

	p.Status = PaymentStatusRefunded
	if reason != "" {
		p.Metadata["refund_reason"] = reason
	}
	p.UpdatedAt = time.Now()

	return nil
}

// IsAuthorized returns true if the payment is authorized
func (p *Payment) IsAuthorized() bool {
	return p.Status == PaymentStatusAuthorized
}

// IsCaptured returns true if the payment is captured
func (p *Payment) IsCaptured() bool {
	return p.Status == PaymentStatusCaptured
}

// IsFailed returns true if the payment has failed
func (p *Payment) IsFailed() bool {
	return p.Status == PaymentStatusFailed
}

// IsCancelled returns true if the payment is cancelled
func (p *Payment) IsCancelled() bool {
	return p.Status == PaymentStatusCancelled
}

// IsRefunded returns true if the payment is refunded
func (p *Payment) IsRefunded() bool {
	return p.Status == PaymentStatusRefunded
}

// CanBeCaptured returns true if the payment can be captured
func (p *Payment) CanBeCaptured() bool {
	return p.Status == PaymentStatusAuthorized
}

// CanBeCancelled returns true if the payment can be cancelled
func (p *Payment) CanBeCancelled() bool {
	return p.Status == PaymentStatusPending || p.Status == PaymentStatusAuthorized
}

// CanBeRefunded returns true if the payment can be refunded
func (p *Payment) CanBeRefunded() bool {
	return p.Status == PaymentStatusCaptured
}
