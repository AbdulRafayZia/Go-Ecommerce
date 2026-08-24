package domain

type OrderStatus string

const (
	OrderStatusPending OrderStatus = "pending"

	OrderStatusAwaitingPayment OrderStatus = "awaiting_payment"

	OrderStatusPaid OrderStatus = "paid"

	OrderStatusFulfilling OrderStatus = "fulfilling"

	OrderStatusShipped OrderStatus = "shipped"

	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"

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


func (s OrderStatus) IsFinal() bool {
	return s == OrderStatusDelivered || s == OrderStatusCancelled || s == OrderStatusFailed
}


func (s OrderStatus) CanTransitionTo(target OrderStatus) bool {
		
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
