package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"gocommerce/services/product/internal/domain"
)

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResp := ErrorResponse{
		Error: Error{
			Code:    code,
			Message: message,
		},
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
	case errors.Is(err, domain.ErrProductNotFound):
		writeJSONError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrProductAlreadyExists):
		writeJSONError(w, http.StatusConflict, "PRODUCT_ALREADY_EXISTS", err.Error())
	case errors.Is(err, domain.ErrCategoryNotFound):
		writeJSONError(w, http.StatusNotFound, "CATEGORY_NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrCategoryAlreadyExists):
		writeJSONError(w, http.StatusConflict, "CATEGORY_ALREADY_EXISTS", err.Error())
	case errors.Is(err, domain.ErrEmptyProductName):
		writeJSONError(w, http.StatusBadRequest, "INVALID_PRODUCT_NAME", err.Error())
	case errors.Is(err, domain.ErrEmptyCategoryName):
		writeJSONError(w, http.StatusBadRequest, "INVALID_CATEGORY_NAME", err.Error())
	case errors.Is(err, domain.ErrInvalidPrice):
		writeJSONError(w, http.StatusBadRequest, "INVALID_PRICE", err.Error())
	case errors.Is(err, domain.ErrInvalidStock):
		writeJSONError(w, http.StatusBadRequest, "INVALID_STOCK", err.Error())
	default:
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
	}
}
