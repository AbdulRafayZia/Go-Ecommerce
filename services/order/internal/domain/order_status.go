package domain

// OrderStatus represents the current status of an order
type OrderStatus string

const (
	// OrderStatusPending is the initial status when an order is created
	OrderStatusPending OrderStatus = "pending"

	// OrderStatusAwaitingPayment indicates the order is waiting for payment
	OrderStatusAwaitingPayment OrderStatus = "awaiting_payment"

	// OrderStatusPaid indicates the order has been paid
	OrderStatusPaid OrderStatus = "paid"

	// OrderStatusFulfilling indicates the order is being prepared/packaged
	OrderStatusFulfilling OrderStatus = "fulfilling"

	// OrderStatusShipped indicates the order has been shipped
	OrderStatusShipped OrderStatus = "shipped"

	// OrderStatusDelivered indicates the order has been delivered
	OrderStatusDelivered OrderStatus = "delivered"

	// OrderStatusCancelled indicates the order has been cancelled
	OrderStatusCancelled OrderStatus = "cancelled"

	// OrderStatusFailed indicates the order failed (payment failed, etc.)
	OrderStatusFailed OrderStatus = "failed"
)

// IsValid checks if the order status is valid
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending,
		OrderStatusAwaitingPayment,
		OrderStatusPaid,
		OrderStatusFulfilling,
		OrderStatusShipped,
		OrderStatusDelivered,
		OrderStatusCancelled,
		OrderStatusFailed:
		return true
	default:
		return false
	}
}

// IsFinal checks if the order status is a final state
func (s OrderStatus) IsFinal() bool {
	return s == OrderStatusDelivered || s == OrderStatusCancelled || s == OrderStatusFailed
}

// CanTransitionTo checks if a transition from current status to target status is valid
func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
	// Cannot transition from final states
	if s.IsFinal() {
		return false
	}

	// Define valid state transitions
	validTransitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending: {
			OrderStatusAwaitingPayment,
			OrderStatusCancelled,
			OrderStatusFailed,
		},
		OrderStatusAwaitingPayment: {
			OrderStatusPaid,
			OrderStatusCancelled,
			OrderStatusFailed,
		},
		OrderStatusPaid: {
			OrderStatusFulfilling,
			OrderStatusCancelled, // Can still cancel after payment
		},
		OrderStatusFulfilling: {
			OrderStatusShipped,
			OrderStatusCancelled, // Can cancel during fulfillment
		},
		OrderStatusShipped: {
			OrderStatusDelivered,
		},
		// Final states have no valid transitions
		OrderStatusDelivered: {},
		OrderStatusCancelled: {},
		OrderStatusFailed:    {},
	}

	allowedStates, exists := validTransitions[s]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == target {
			return true
		}
	}

	return false
}
