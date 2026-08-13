package middleware

import (
	"context"
	"net/http"

	"gocommerce/pkg/logger"

	"github.com/google/uuid"
)

const (
	// HTTPCorrelationIDHeader is the HTTP header for correlation ID
	HTTPCorrelationIDHeader = "X-Correlation-ID"
)

// HTTPCorrelationID extracts or generates a correlation ID for HTTP requests
func HTTPCorrelationID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract correlation ID from header or generate new one
			correlationID := r.Header.Get(HTTPCorrelationIDHeader)
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			// Add to context
			ctx := context.WithValue(r.Context(), logger.CorrelationIDKey, correlationID)

			// Add to response header
			w.Header().Set(HTTPCorrelationIDHeader, correlationID)

			// Process request with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetCorrelationID retrieves the correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(logger.CorrelationIDKey).(string); ok {
		return correlationID
	}
	return ""
}
