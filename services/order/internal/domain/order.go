package domain

import (
	"strings"
	"time"
)

// Order represents an order (aggregate root)
// This is the main domain entity that contains order items and manages state transitions
type Order struct {
	ID             string
	UserID         string
	Status         OrderStatus
	Items          []*OrderItem
	TotalCents     int
	IdempotencyKey string
	PaymentID      *string
	TrackingNumber *string
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Domain events collected during the lifecycle of this aggregate
	// These will be published through the outbox pattern
	events []DomainEvent
}

// NewOrder creates a new order with validation
func NewOrder(userID, idempotencyKey string, items []*OrderItem) (*Order, error) {
	// Validate user ID
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidUserID
	}

	// Validate idempotency key
	if strings.TrimSpace(idempotencyKey) == "" {
		return nil, ErrInvalidIdempotencyKey
	}

	// Validate items
	if len(items) == 0 {
		return nil, ErrEmptyOrder
	}

	// Calculate total
	total := 0
	for _, item := range items {
		total += item.Subtotal()
	}

	now := time.Now()
	order := &Order{
		UserID:         userID,
		Status:         OrderStatusPending,
		Items:          items,
		TotalCents:     total,
		IdempotencyKey: idempotencyKey,
		CreatedAt:      now,
		UpdatedAt:      now,
		events:         make([]DomainEvent, 0),
	}

	// Record domain event
	order.recordEvent(OrderCreatedEvent{
		OrderID:        order.ID,
		UserID:         order.UserID,
		Items:          orderItemsToValue(items),
		TotalCents:     order.TotalCents,
		IdempotencyKey: order.IdempotencyKey,
		CreatedAt:      order.CreatedAt,
	})

	return order, nil
}

// MarkAwaitingPayment transitions the order to awaiting payment status
func (o *Order) MarkAwaitingPayment() error {
	return o.transitionTo(OrderStatusAwaitingPayment)
}

// MarkPaid marks the order as paid and records the payment ID
func (o *Order) MarkPaid(paymentID string) error {
	if o.Status == OrderStatusPaid {
		return ErrOrderAlreadyPaid
	}

	if err := o.transitionTo(OrderStatusPaid); err != nil {
		return err
	}

	o.PaymentID = &paymentID

	// Record domain event
	o.recordEvent(OrderPaidEvent{
		OrderID:   o.ID,
		UserID:    o.UserID,
		PaymentID: paymentID,
		PaidAt:    time.Now(),
	})

	return nil
}

// MarkFulfilling transitions the order to fulfilling status
func (o *Order) MarkFulfilling() error {
	if o.Status != OrderStatusPaid {
		return ErrOrderNotPaid
	}

	return o.transitionTo(OrderStatusFulfilling)
}

// MarkShipped marks the order as shipped with optional tracking number
func (o *Order) MarkShipped(trackingNumber *string) error {
	if err := o.transitionTo(OrderStatusShipped); err != nil {
		return err
	}

	o.TrackingNumber = trackingNumber

	// Record domain event
	o.recordEvent(OrderShippedEvent{
		OrderID:        o.ID,
		UserID:         o.UserID,
		TrackingNumber: trackingNumber,
		ShippedAt:      time.Now(),
	})

	return nil
}

// MarkDelivered marks the order as delivered
func (o *Order) MarkDelivered() error {
	if o.Status == OrderStatusDelivered {
		return ErrOrderAlreadyDelivered
	}

	if err := o.transitionTo(OrderStatusDelivered); err != nil {
		return err
	}

	// Record domain event
	o.recordEvent(OrderDeliveredEvent{
		OrderID:     o.ID,
		UserID:      o.UserID,
		DeliveredAt: time.Now(),
	})

	return nil
}

// Cancel cancels the order with a reason
func (o *Order) Cancel(reason string) error {
	if o.Status == OrderStatusCancelled {
		return ErrOrderAlreadyCancelled
	}

	if o.Status == OrderStatusDelivered {
		return ErrCannotCancelOrder
	}

	if err := o.transitionTo(OrderStatusCancelled); err != nil {
		return err
	}

	// Record domain event
	o.recordEvent(OrderCancelledEvent{
		OrderID:     o.ID,
		UserID:      o.UserID,
		Reason:      reason,
		CancelledAt: time.Now(),
	})

	return nil
}

// MarkFailed marks the order as failed with a reason
func (o *Order) MarkFailed(reason string) error {
	if err := o.transitionTo(OrderStatusFailed); err != nil {
		return err
	}

	// Record domain event
	o.recordEvent(OrderFailedEvent{
		OrderID:  o.ID,
		UserID:   o.UserID,
		Reason:   reason,
		FailedAt: time.Now(),
	})

	return nil
}

// transitionTo performs a state transition with validation
func (o *Order) transitionTo(newStatus OrderStatus) error {
	if !o.Status.CanTransitionTo(newStatus) {
		return ErrInvalidStateTransition
	}

	o.Status = newStatus
	o.UpdatedAt = time.Now()

	return nil
}

// GetEvents returns all domain events collected by this aggregate
func (o *Order) GetEvents() []DomainEvent {
	return o.events
}

// ClearEvents clears all domain events (called after events are published)
func (o *Order) ClearEvents() {
	o.events = make([]DomainEvent, 0)
}

// recordEvent adds a domain event to the aggregate's event collection
func (o *Order) recordEvent(event DomainEvent) {
	o.events = append(o.events, event)
}

// IsPaid checks if the order has been paid
func (o *Order) IsPaid() bool {
	return o.Status == OrderStatusPaid ||
		o.Status == OrderStatusFulfilling ||
		o.Status == OrderStatusShipped ||
		o.Status == OrderStatusDelivered
}

// IsCancellable checks if the order can be cancelled
func (o *Order) IsCancellable() bool {
	return o.Status != OrderStatusDelivered &&
		o.Status != OrderStatusCancelled &&
		o.Status != OrderStatusFailed
}

// IsFinal checks if the order is in a final state
func (o *Order) IsFinal() bool {
	return o.Status.IsFinal()
}

// orderItemsToValue converts order item pointers to values for events
func orderItemsToValue(items []*OrderItem) []OrderItem {
	result := make([]OrderItem, len(items))
	for i, item := range items {
		result[i] = *item
	}
	return result
}
