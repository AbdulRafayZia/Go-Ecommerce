package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"gocommerce/services/inventory/internal/domain"
)

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResp := Error{
		Error:   code,
		Message: message,
	}

	json.NewEncoder(w).Encode(errResp)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// handleError converts domain errors to HTTP responses
func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrStockNotFound):
		writeJSONError(w, http.StatusNotFound, "STOCK_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrReservationNotFound):
		writeJSONError(w, http.StatusNotFound, "RESERVATION_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrInsufficientStock):
		writeJSONError(w, http.StatusBadRequest, "INSUFFICIENT_STOCK", err.Error())
	case errors.Is(err, domain.ErrInvalidQuantity):
		writeJSONError(w, http.StatusBadRequest, "INVALID_QUANTITY", err.Error())
	case errors.Is(err, domain.ErrInvalidProductID):
		writeJSONError(w, http.StatusBadRequest, "INVALID_PRODUCT_ID", err.Error())
	case errors.Is(err, domain.ErrInvalidOrderID):
		writeJSONError(w, http.StatusBadRequest, "INVALID_ORDER_ID", err.Error())
	case errors.Is(err, domain.ErrNegativeStock):
		writeJSONError(w, http.StatusBadRequest, "NEGATIVE_STOCK", err.Error())
	case errors.Is(err, domain.ErrReservationAlreadyFulfilled):
		writeJSONError(w, http.StatusConflict, "RESERVATION_ALREADY_FULFILLED", err.Error())
	case errors.Is(err, domain.ErrReservationAlreadyCancelled):
		writeJSONError(w, http.StatusConflict, "RESERVATION_ALREADY_CANCELLED", err.Error())
	default:
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
	}
}
