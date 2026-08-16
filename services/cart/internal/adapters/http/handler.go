package api

import (
	"encoding/json"
	"net/http"
	"time"

	"gocommerce/services/cart/internal/application"
)

// CartHandler implements the ServerInterface from the generated code
type CartHandler struct {
	cartService *application.CartService
}

// NewCartHandler creates a new cart HTTP handler
func NewCartHandler(cartService *application.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// HealthCheck implements the health check endpoint
// (GET /health)
func (h *CartHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Service:   "cart-service",
		Timestamp: ptrTime(time.Now()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetCart retrieves a user's cart
// (GET /carts/{userId})
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request, userId string) {
	cart, err := h.cartService.GetCart(r.Context(), userId)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	response := domainCartToAPI(cart)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ClearCart removes all items from a user's cart
// (DELETE /carts/{userId})
func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request, userId string) {
	if err := h.cartService.ClearCart(r.Context(), userId); err != nil {
		mapDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddItem adds an item to the cart
// (POST /carts/{userId}/items)
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request, userId string) {
	var req AddItemRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Convert API request to application DTO
	dto := application.AddItemDTO{
		ProductID:  req.ProductId,
		Name:       req.Name,
		PriceCents: req.PriceCents,
		Quantity:   req.Quantity,
		ImageURL:   req.ImageUrl,
	}

	cart, err := h.cartService.AddItem(r.Context(), userId, dto)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	response := domainCartToAPI(cart)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateItemQuantity updates the quantity of an item in the cart
// (PUT /carts/{userId}/items/{productId})
func (h *CartHandler) UpdateItemQuantity(w http.ResponseWriter, r *http.Request, userId string, productId string) {
	var req UpdateQuantityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	cart, err := h.cartService.UpdateItemQuantity(r.Context(), userId, productId, req.Quantity)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	response := domainCartToAPI(cart)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RemoveItem removes an item from the cart
// (DELETE /carts/{userId}/items/{productId})
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request, userId string, productId string) {
	cart, err := h.cartService.RemoveItem(r.Context(), userId, productId)
	if err != nil {
		mapDomainError(w, err)
		return
	}

	response := domainCartToAPI(cart)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func ptrTime(t time.Time) *time.Time {
	return &t
}
