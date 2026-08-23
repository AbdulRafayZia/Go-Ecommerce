package http

import (
	"encoding/json"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"gocommerce/services/inventory/internal/application"
)

// InventoryHandler implements the ServerInterface
type InventoryHandler struct {
	inventoryService *application.InventoryService
}

// NewInventoryHandler creates a new HTTP handler
func NewInventoryHandler(inventoryService *application.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// HealthCheck implements the health check endpoint
// (GET /health)
func (h *InventoryHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	response := HealthResponse{
		Status:    "healthy",
		Service:   "inventory-service",
		Timestamp: &now,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetStock retrieves stock information for a product
// (GET /stocks/{productId})
func (h *InventoryHandler) GetStock(w http.ResponseWriter, r *http.Request, productId openapi_types.UUID) {
	productIDStr := uuidToString(productId)

	stock, err := h.inventoryService.GetStock(r.Context(), productIDStr)
	if err != nil {
		handleError(w, err)
		return
	}

	response := domainStockToResponse(stock)
	writeJSON(w, http.StatusOK, response)
}

// AddStock adds stock to a product (restocking operation)
// (POST /stocks/{productId}/add)
func (h *InventoryHandler) AddStock(w http.ResponseWriter, r *http.Request, productId openapi_types.UUID) {
	var req AddStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	productIDStr := uuidToString(productId)

	if err := h.inventoryService.AddStock(r.Context(), productIDStr, req.Quantity); err != nil {
		handleError(w, err)
		return
	}

	// Get updated stock to return
	stock, err := h.inventoryService.GetStock(r.Context(), productIDStr)
	if err != nil {
		handleError(w, err)
		return
	}

	response := domainStockToResponse(stock)
	writeJSON(w, http.StatusOK, response)
}

// SetStock sets the available stock for a product (inventory adjustment)
// (POST /stocks/{productId}/set)
func (h *InventoryHandler) SetStock(w http.ResponseWriter, r *http.Request, productId openapi_types.UUID) {
	var req SetStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	productIDStr := uuidToString(productId)

	if err := h.inventoryService.SetStock(r.Context(), productIDStr, req.Quantity); err != nil {
		handleError(w, err)
		return
	}

	// Get updated stock to return
	stock, err := h.inventoryService.GetStock(r.Context(), productIDStr)
	if err != nil {
		handleError(w, err)
		return
	}

	response := domainStockToResponse(stock)
	writeJSON(w, http.StatusOK, response)
}

// ListLowStock returns products with stock below the specified threshold
// (GET /stocks/low)
func (h *InventoryHandler) ListLowStock(w http.ResponseWriter, r *http.Request, params ListLowStockParams) {
	threshold := 10 // Default threshold
	if params.Threshold != nil {
		threshold = *params.Threshold
	}

	stocks, err := h.inventoryService.ListLowStock(r.Context(), threshold)
	if err != nil {
		handleError(w, err)
		return
	}

	// Convert to response format
	products := make([]StockResponse, len(stocks))
	for i, stock := range stocks {
		products[i] = domainStockToResponse(stock)
	}

	response := LowStockResponse{
		Products:  products,
		Threshold: threshold,
		Count:     len(products),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetReservationsByOrder retrieves all stock reservations for a specific order
// (GET /reservations/{orderId})
func (h *InventoryHandler) GetReservationsByOrder(w http.ResponseWriter, r *http.Request, orderId openapi_types.UUID) {
	orderIDStr := uuidToString(orderId)

	reservations, err := h.inventoryService.GetReservationsByOrderID(r.Context(), orderIDStr)
	if err != nil {
		handleError(w, err)
		return
	}

	// Convert to response format
	items := make([]ReservationItem, len(reservations))
	for i, reservation := range reservations {
		items[i] = domainReservationToItem(reservation)
	}

	response := ReservationsResponse{
		OrderId:      orderId,
		Reservations: items,
		Count:        len(items),
	}

	writeJSON(w, http.StatusOK, response)
}
