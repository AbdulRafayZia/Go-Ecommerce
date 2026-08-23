package kafka

import "time"

// PaymentEvent represents the base structure for payment events
type PaymentEvent struct {
	EventType string                 `json:"event_type"`
	PaymentID string                 `json:"payment_id"`
	OrderID   string                 `json:"order_id"`
	Amount    float64                `json:"amount"`
	Currency  string                 `json:"currency"`
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

const (
	// Event types
	EventTypePaymentAuthorized = "payment.authorized"
	EventTypePaymentCaptured   = "payment.captured"
	EventTypePaymentFailed     = "payment.failed"
	EventTypePaymentCancelled  = "payment.cancelled"
	EventTypePaymentRefunded   = "payment.refunded"
)

const (
	// Kafka topics
	TopicPaymentAuthorized = "payment.authorized"
	TopicPaymentCaptured   = "payment.captured"
	TopicPaymentFailed     = "payment.failed"
	TopicPaymentCancelled  = "payment.cancelled"
	TopicPaymentRefunded   = "payment.refunded"
)

// NewPaymentEvent creates a new payment event
func NewPaymentEvent(eventType, paymentID, orderID string, amount float64, currency, status string) *PaymentEvent {
	return &PaymentEvent{
		EventType: eventType,
		PaymentID: paymentID,
		OrderID:   orderID,
		Amount:    amount,
		Currency:  currency,
		Status:    status,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}
