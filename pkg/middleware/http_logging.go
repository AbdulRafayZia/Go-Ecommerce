package middleware

import (
	"net/http"
	"time"

	"gocommerce/pkg/logger"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// HTTPLogging logs HTTP requests and responses
func HTTPLogging(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer
			wrapped := newResponseWriter(w)

			// Get correlation ID if set
			correlationID := GetCorrelationID(r.Context())
			logWithCtx := log
			if correlationID != "" {
				logWithCtx = log.WithField("correlation_id", correlationID)
			}

			// Log request
			logWithCtx.WithFields(map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			}).Info("HTTP request started")

			// Process request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Log response
			logEntry := logWithCtx.WithFields(map[string]interface{}{
				"method":       r.Method,
				"path":         r.URL.Path,
				"status":       wrapped.statusCode,
				"duration_ms":  duration.Milliseconds(),
				"bytes_written": wrapped.written,
			})

			if wrapped.statusCode >= 500 {
				logEntry.Error("HTTP request failed with server error")
			} else if wrapped.statusCode >= 400 {
				logEntry.Warn("HTTP request failed with client error")
			} else {
				logEntry.Info("HTTP request completed")
			}
		})
	}
}
