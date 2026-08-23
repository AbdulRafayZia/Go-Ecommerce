package domain

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	// PaymentStatusPending indicates the payment is pending
	PaymentStatusPending PaymentStatus = "pending"

	// PaymentStatusAuthorized indicates the payment has been authorized (funds reserved)
	PaymentStatusAuthorized PaymentStatus = "authorized"

	// PaymentStatusCaptured indicates the payment has been captured (funds charged)
	PaymentStatusCaptured PaymentStatus = "captured"

	// PaymentStatusFailed indicates the payment has failed
	PaymentStatusFailed PaymentStatus = "failed"

	// PaymentStatusCancelled indicates the payment has been cancelled
	PaymentStatusCancelled PaymentStatus = "cancelled"

	// PaymentStatusRefunded indicates the payment has been refunded
	PaymentStatusRefunded PaymentStatus = "refunded"
)

// CanTransitionTo checks if a payment can transition from current status to target status
func (s PaymentStatus) CanTransitionTo(target PaymentStatus) bool {
	transitions := map[PaymentStatus][]PaymentStatus{
		PaymentStatusPending: {
			PaymentStatusAuthorized,
			PaymentStatusFailed,
			PaymentStatusCancelled,
		},
		PaymentStatusAuthorized: {
			PaymentStatusCaptured,
			PaymentStatusFailed,
			PaymentStatusCancelled,
		},
		PaymentStatusCaptured: {
			PaymentStatusRefunded,
		},
		PaymentStatusFailed:    {},
		PaymentStatusCancelled: {},
		PaymentStatusRefunded:  {},
	}

	allowedTargets, exists := transitions[s]
	if !exists {
		return false
	}

	for _, allowed := range allowedTargets {
		if allowed == target {
			return true
		}
	}

	return false
}

// IsFinal returns true if the status is a terminal state
func (s PaymentStatus) IsFinal() bool {
	return s == PaymentStatusFailed ||
		s == PaymentStatusCancelled ||
		s == PaymentStatusRefunded ||
		s == PaymentStatusCaptured
}

// IsValid checks if the payment status is valid
func (s PaymentStatus) IsValid() bool {
	switch s {
	case PaymentStatusPending,
		PaymentStatusAuthorized,
		PaymentStatusCaptured,
		PaymentStatusFailed,
		PaymentStatusCancelled,
		PaymentStatusRefunded:
		return true
	default:
		return false
	}
}
