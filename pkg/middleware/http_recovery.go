package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"gocommerce/pkg/logger"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HTTPRecovery recovers from panics in HTTP handlers
func HTTPRecovery(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					log.WithCorrelationID(r.Context()).WithFields(map[string]interface{}{
						"method":      r.Method,
						"path":        r.URL.Path,
						"panic":       err,
						"stack_trace": string(debug.Stack()),
					}).Error("panic recovered in HTTP handler")

					// Return internal server error
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					response := ErrorResponse{
						Error: ErrorDetail{
							Code:    "INTERNAL_SERVER_ERROR",
							Message: "An internal server error occurred",
						},
					}

					json.NewEncoder(w).Encode(response)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// WriteError writes a standard error response
func WriteError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// WriteErrorWithDetails writes an error response with additional details
func WriteErrorWithDetails(w http.ResponseWriter, statusCode int, code, message string, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	json.NewEncoder(w).Encode(response)
}
