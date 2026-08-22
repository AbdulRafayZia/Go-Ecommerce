package kafka

import "time"

// OrderCreatedEvent represents an order.created event from the Order Service
type OrderCreatedEvent struct {
	OrderID   string             `json:"order_id"`
	UserID    string             `json:"user_id"`
	Items     []OrderCreatedItem `json:"items"`
	CreatedAt time.Time          `json:"created_at"`
}

// OrderCreatedItem represents an item in an order.created event
type OrderCreatedItem struct {
	ProductID    string `json:"product_id"`
	ProductName  string `json:"product_name"`
	Quantity     int    `json:"quantity"`
	PriceCents   int    `json:"price_cents"`
	SubtotalCents int   `json:"subtotal_cents"`
}

// OrderPaidEvent represents an order.paid event from the Order Service
type OrderPaidEvent struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	PaymentID string    `json:"payment_id"`
	PaidAt    time.Time `json:"paid_at"`
}

// OrderCancelledEvent represents an order.cancelled event from the Order Service
type OrderCancelledEvent struct {
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}
