package api

import (
	"encoding/json"
	"net/http"

	"gocommerce/services/cart/internal/domain"
)

// writeError writes an error response in JSON format
func writeError(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   errorCode,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// mapDomainError maps domain errors to HTTP status codes and error responses
func mapDomainError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrCartNotFound:
		writeError(w, http.StatusNotFound, "CART_NOT_FOUND", "Cart not found for user")
	case domain.ErrItemNotFound:
		writeError(w, http.StatusNotFound, "ITEM_NOT_FOUND", "Item not found in cart")
	case domain.ErrInvalidQuantity:
		writeError(w, http.StatusBadRequest, "INVALID_QUANTITY", "Quantity must be greater than zero")
	case domain.ErrInvalidUserID:
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "User ID cannot be empty")
	case domain.ErrInvalidProductID:
		writeError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Product ID cannot be empty")
	case domain.ErrInvalidPrice:
		writeError(w, http.StatusBadRequest, "INVALID_PRICE", "Price must be greater than or equal to zero")
	case domain.ErrEmptyCart:
		writeError(w, http.StatusBadRequest, "EMPTY_CART", "Cart is empty")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
	}
}
