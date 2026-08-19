package domain

import "time"

// DomainEvent is the base interface for all domain events
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

// OrderCreatedEvent is published when a new order is created
type OrderCreatedEvent struct {
	OrderID        string
	UserID         string
	Items          []OrderItem
	TotalCents     int
	IdempotencyKey string
	CreatedAt      time.Time
}

func (e OrderCreatedEvent) EventType() string   { return "order.created" }
func (e OrderCreatedEvent) AggregateID() string { return e.OrderID }
func (e OrderCreatedEvent) OccurredAt() time.Time { return e.CreatedAt }

// OrderPaidEvent is published when an order payment is confirmed
type OrderPaidEvent struct {
	OrderID   string
	UserID    string
	PaymentID string
	PaidAt    time.Time
}

func (e OrderPaidEvent) EventType() string   { return "order.paid" }
func (e OrderPaidEvent) AggregateID() string { return e.OrderID }
func (e OrderPaidEvent) OccurredAt() time.Time { return e.PaidAt }

// OrderShippedEvent is published when an order is shipped
type OrderShippedEvent struct {
	OrderID       string
	UserID        string
	TrackingNumber *string
	ShippedAt     time.Time
}

func (e OrderShippedEvent) EventType() string   { return "order.shipped" }
func (e OrderShippedEvent) AggregateID() string { return e.OrderID }
func (e OrderShippedEvent) OccurredAt() time.Time { return e.ShippedAt }

// OrderDeliveredEvent is published when an order is delivered
type OrderDeliveredEvent struct {
	OrderID     string
	UserID      string
	DeliveredAt time.Time
}

func (e OrderDeliveredEvent) EventType() string   { return "order.delivered" }
func (e OrderDeliveredEvent) AggregateID() string { return e.OrderID }
func (e OrderDeliveredEvent) OccurredAt() time.Time { return e.DeliveredAt }

// OrderCancelledEvent is published when an order is cancelled
type OrderCancelledEvent struct {
	OrderID     string
	UserID      string
	Reason      string
	CancelledAt time.Time
}

func (e OrderCancelledEvent) EventType() string   { return "order.cancelled" }
func (e OrderCancelledEvent) AggregateID() string { return e.OrderID }
func (e OrderCancelledEvent) OccurredAt() time.Time { return e.CancelledAt }

// OrderFailedEvent is published when an order fails
type OrderFailedEvent struct {
	OrderID  string
	UserID   string
	Reason   string
	FailedAt time.Time
}

func (e OrderFailedEvent) EventType() string   { return "order.failed" }
func (e OrderFailedEvent) AggregateID() string { return e.OrderID }
func (e OrderFailedEvent) OccurredAt() time.Time { return e.FailedAt }
