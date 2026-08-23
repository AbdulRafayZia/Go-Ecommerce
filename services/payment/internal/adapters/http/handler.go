package http

import (
	"encoding/json"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"gocommerce/services/payment/internal/application"
	"gocommerce/services/payment/internal/domain"
	"gocommerce/services/payment/internal/ports"
)

// PaymentHandler implements the ServerInterface
type PaymentHandler struct {
	paymentService *application.PaymentService
}

// NewPaymentHandler creates a new HTTP handler
func NewPaymentHandler(paymentService *application.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// HealthCheck implements the health check endpoint
// (GET /health)
func (h *PaymentHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	response := HealthResponse{
		Status:    "healthy",
		Service:   "payment-service",
		Timestamp: &now,
	}

	writeJSON(w, http.StatusOK, response)
}

// CreatePayment creates a new payment (idempotent)
// (POST /payments)
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Generate idempotency key if not provided
	idempotencyKey := "default"
	if req.IdempotencyKey != nil {
		idempotencyKey = *req.IdempotencyKey
	} else {
		// Use order_id as idempotency key if not provided
		idempotencyKey = uuidToString(req.OrderId)
	}

	// Determine payment method
	method := domain.PaymentMethodCard
	if req.PaymentMethod != nil {
		switch *req.PaymentMethod {
		case BankTransfer:
			method = domain.PaymentMethodBankTransfer
		case Wallet:
			method = domain.PaymentMethodWallet
		}
	}

	// Create payment
	payment, err := h.paymentService.CreatePayment(
		r.Context(),
		uuidToString(req.OrderId),
		req.Amount,
		req.Currency,
		idempotencyKey,
		method,
	)

	if err != nil {
		handleError(w, err)
		return
	}

	// Determine status code (201 for new, 200 for existing)
	statusCode := http.StatusCreated
	if payment.CreatedAt.Before(time.Now().Add(-1 * time.Second)) {
		statusCode = http.StatusOK // Idempotent response
	}

	response := domainPaymentToResponse(payment)
	writeJSON(w, statusCode, response)
}

// GetPayment retrieves a payment by ID
// (GET /payments/{paymentId})
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request, paymentId openapi_types.UUID) {
	payment, err := h.paymentService.GetPayment(r.Context(), uuidToString(paymentId))
	if err != nil {
		handleError(w, err)
		return
	}

	response := domainPaymentToResponse(payment)
	writeJSON(w, http.StatusOK, response)
}

// ListPayments lists payments with optional filtering
// (GET /payments)
func (h *PaymentHandler) ListPayments(w http.ResponseWriter, r *http.Request, params ListPaymentsParams) {
	filter := ports.PaymentFilter{
		Limit:  20,
		Offset: 0,
	}

	if params.OrderId != nil {
		orderID := uuidToString(*params.OrderId)
		filter.OrderID = &orderID
	}

	if params.Status != nil {
		status := domain.PaymentStatus(*params.Status)
		filter.Status = &status
	}

	if params.Limit != nil {
		filter.Limit = *params.Limit
	}

	if params.Offset != nil {
		filter.Offset = *params.Offset
	}

	payments, total, err := h.paymentService.ListPayments(r.Context(), filter)
	if err != nil {
		handleError(w, err)
		return
	}

	// Convert to response format
	paymentResponses := make([]PaymentResponse, len(payments))
	for i, payment := range payments {
		paymentResponses[i] = domainPaymentToResponse(payment)
	}

	response := PaymentListResponse{
		Payments: paymentResponses,
		Total:    total,
		Limit:    filter.Limit,
		Offset:   filter.Offset,
	}

	writeJSON(w, http.StatusOK, response)
}

// CapturePayment captures an authorized payment
// (POST /payments/{paymentId}/capture)
func (h *PaymentHandler) CapturePayment(w http.ResponseWriter, r *http.Request, paymentId openapi_types.UUID) {
	payment, err := h.paymentService.CapturePayment(r.Context(), uuidToString(paymentId))
	if err != nil {
		handleError(w, err)
		return
	}

	response := domainPaymentToResponse(payment)
	writeJSON(w, http.StatusOK, response)
}

// CancelPayment cancels a pending or authorized payment
// (POST /payments/{paymentId}/cancel)
func (h *PaymentHandler) CancelPayment(w http.ResponseWriter, r *http.Request, paymentId openapi_types.UUID) {
	payment, err := h.paymentService.CancelPayment(r.Context(), uuidToString(paymentId))
	if err != nil {
		handleError(w, err)
		return
	}

	response := domainPaymentToResponse(payment)
	writeJSON(w, http.StatusOK, response)
}

// RefundPayment refunds a captured payment
// (POST /payments/{paymentId}/refund)
func (h *PaymentHandler) RefundPayment(w http.ResponseWriter, r *http.Request, paymentId openapi_types.UUID) {
	var req RefundRequest
	reason := ""
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if req.Reason != nil {
				reason = *req.Reason
			}
		}
	}

	payment, err := h.paymentService.RefundPayment(r.Context(), uuidToString(paymentId), reason)
	if err != nil {
		handleError(w, err)
		return
	}

	response := domainPaymentToResponse(payment)
	writeJSON(w, http.StatusOK, response)
}

// HandleWebhook handles payment provider webhooks
// (POST /webhooks/payment-provider)
func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var event WebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_WEBHOOK", "Invalid webhook payload")
		return
	}

	// Extract data as map
	data := make(map[string]interface{})
	if event.Data != nil {
		data = *event.Data
	}

	// Handle webhook
	err := h.paymentService.HandleWebhook(r.Context(), string(event.EventType), event.PaymentId, data)
	if err != nil {
		handleError(w, err)
		return
	}

	response := WebhookResponse{
		Received: true,
	}

	writeJSON(w, http.StatusOK, response)
}
