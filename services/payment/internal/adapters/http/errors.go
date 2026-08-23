package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"gocommerce/services/payment/internal/domain"
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
	case errors.Is(err, domain.ErrPaymentNotFound):
		writeJSONError(w, http.StatusNotFound, "PAYMENT_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrPaymentAlreadyExists):
		writeJSONError(w, http.StatusConflict, "PAYMENT_ALREADY_EXISTS", err.Error())
	case errors.Is(err, domain.ErrInvalidAmount):
		writeJSONError(w, http.StatusBadRequest, "INVALID_AMOUNT", err.Error())
	case errors.Is(err, domain.ErrInvalidCurrency):
		writeJSONError(w, http.StatusBadRequest, "INVALID_CURRENCY", err.Error())
	case errors.Is(err, domain.ErrInvalidOrderID):
		writeJSONError(w, http.StatusBadRequest, "INVALID_ORDER_ID", err.Error())
	case errors.Is(err, domain.ErrInvalidStateTransition):
		writeJSONError(w, http.StatusBadRequest, "INVALID_STATE_TRANSITION", err.Error())
	case errors.Is(err, domain.ErrPaymentNotAuthorized):
		writeJSONError(w, http.StatusBadRequest, "PAYMENT_NOT_AUTHORIZED", err.Error())
	case errors.Is(err, domain.ErrPaymentNotCaptured):
		writeJSONError(w, http.StatusBadRequest, "PAYMENT_NOT_CAPTURED", err.Error())
	case errors.Is(err, domain.ErrPaymentAlreadyCaptured):
		writeJSONError(w, http.StatusConflict, "PAYMENT_ALREADY_CAPTURED", err.Error())
	case errors.Is(err, domain.ErrPaymentAlreadyCancelled):
		writeJSONError(w, http.StatusConflict, "PAYMENT_ALREADY_CANCELLED", err.Error())
	case errors.Is(err, domain.ErrPaymentAlreadyRefunded):
		writeJSONError(w, http.StatusConflict, "PAYMENT_ALREADY_REFUNDED", err.Error())
	case errors.Is(err, domain.ErrProviderError):
		writeJSONError(w, http.StatusBadGateway, "PROVIDER_ERROR", err.Error())
	default:
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
	}
}
