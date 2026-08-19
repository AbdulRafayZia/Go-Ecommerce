package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gocommerce/services/order/internal/application"
	"gocommerce/services/order/internal/domain"
)

// OrderHandler implements the ServerInterface from the generated code
type OrderHandler struct {
	orderService *application.OrderService
}

// NewOrderHandler creates a new order HTTP handler
func NewOrderHandler(orderService *application.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// HealthCheck implements the health check endpoint
// (GET /health)
func (h *OrderHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Service:   "order-service",
		Timestamp: ptrTime(time.Now()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateOrder creates a new order
// (POST /orders)
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Convert to DTO
	dto := createOrderRequestToDTO(req)

	// Create order
	order, err := h.orderService.CreateOrder(r.Context(), dto)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	// Convert to API response
	response := domainOrderToAPI(order)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetOrder retrieves an order by ID
// (GET /orders/{orderId})
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request, orderId OrderIdParam) {
	orderIDStr := uuidToString(orderId)
	order, err := h.orderService.GetOrder(r.Context(), orderIDStr)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	response := domainOrderToAPI(order)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ListOrders retrieves orders with pagination and filtering
// (GET /orders)
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request, params ListOrdersParams) {
	// Parse filters
	var userID *string
	if params.UserId != nil {
		userID = params.UserId
	}

	var status *domain.OrderStatus
	if params.Status != nil {
		s := domain.OrderStatus(*params.Status)
		status = &s
	}

	page := 1
	if params.Page != nil {
		page = *params.Page
	}

	pageSize := 20
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	// List orders
	orders, total, err := h.orderService.ListOrders(r.Context(), userID, status, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list orders")
		return
	}

	// Convert to API response
	apiOrders := make([]Order, 0, len(orders))
	for _, order := range orders {
		apiOrders = append(apiOrders, domainOrderToAPI(order))
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	response := OrderListResponse{
		Orders:     apiOrders,
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: &totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateOrderStatus updates the order status
// (PUT /orders/{orderId}/status)
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request, orderId OrderIdParam) {
	var req UpdateStatusRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Convert to DTO
	dto := updateStatusRequestToDTO(req)

	// Update status
	orderIDStr := uuidToString(orderId)
	order, err := h.orderService.UpdateOrderStatus(r.Context(), orderIDStr, dto)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	// Convert to API response
	response := domainOrderToAPI(order)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CancelOrder cancels an order
// (POST /orders/{orderId}/cancel)
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request, orderId OrderIdParam) {
	var req CancelOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Cancel order
	orderIDStr := uuidToString(orderId)
	order, err := h.orderService.CancelOrder(r.Context(), orderIDStr, req.Reason)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	// Convert to API response
	response := domainOrderToAPI(order)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func ptrTime(t time.Time) *time.Time {
	return &t
}

func ptrInt(i int) *int {
	return &i
}

func ptrString(s string) *string {
	return &s
}

func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return val
}
