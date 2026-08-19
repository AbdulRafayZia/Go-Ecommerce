package api

import (
	"encoding/json"
	"net/http"

	"gocommerce/services/order/internal/domain"
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
	case domain.ErrOrderNotFound:
		writeError(w, http.StatusNotFound, "ORDER_NOT_FOUND", "Order not found")
	case domain.ErrInvalidStateTransition:
		writeError(w, http.StatusBadRequest, "INVALID_STATE_TRANSITION", "Invalid state transition")
	case domain.ErrEmptyOrder:
		writeError(w, http.StatusBadRequest, "EMPTY_ORDER", "Order must have at least one item")
	case domain.ErrInvalidUserID:
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "User ID cannot be empty")
	case domain.ErrInvalidProductID:
		writeError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", "Product ID cannot be empty")
	case domain.ErrInvalidQuantity:
		writeError(w, http.StatusBadRequest, "INVALID_QUANTITY", "Quantity must be greater than zero")
	case domain.ErrInvalidPrice:
		writeError(w, http.StatusBadRequest, "INVALID_PRICE", "Price must be greater than or equal to zero")
	case domain.ErrOrderAlreadyPaid:
		writeError(w, http.StatusBadRequest, "ORDER_ALREADY_PAID", "Order is already paid")
	case domain.ErrOrderNotPaid:
		writeError(w, http.StatusBadRequest, "ORDER_NOT_PAID", "Order has not been paid")
	case domain.ErrOrderAlreadyCancelled:
		writeError(w, http.StatusBadRequest, "ORDER_ALREADY_CANCELLED", "Order has been cancelled")
	case domain.ErrOrderAlreadyDelivered:
		writeError(w, http.StatusBadRequest, "ORDER_ALREADY_DELIVERED", "Order has already been delivered")
	case domain.ErrCannotCancelOrder:
		writeError(w, http.StatusBadRequest, "CANNOT_CANCEL_ORDER", "Order cannot be cancelled in current state")
	case domain.ErrInvalidIdempotencyKey:
		writeError(w, http.StatusBadRequest, "INVALID_IDEMPOTENCY_KEY", "Idempotency key cannot be empty")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
	}
}
