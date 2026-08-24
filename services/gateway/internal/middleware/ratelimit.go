package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// RequestsPerMinute is the maximum number of requests per minute per IP
	RequestsPerMinute int
	// BurstSize is the maximum burst size
	BurstSize int
}

// DefaultRateLimitConfig returns the default rate limit configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 100,
		BurstSize:         20,
	}
}

// RateLimit creates a rate limiting middleware based on IP address
func RateLimit(cfg RateLimitConfig) func(http.Handler) http.Handler {
	return httprate.Limit(
		cfg.RequestsPerMinute,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			writeJSONError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests, please try again later")
		}),
	)
}

// RateLimitByUser creates a rate limiting middleware based on authenticated user
func RateLimitByUser(cfg RateLimitConfig) func(http.Handler) http.Handler {
	return httprate.Limit(
		cfg.RequestsPerMinute,
		time.Minute,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			// Try to get user ID from context (set by auth middleware)
			userID, ok := GetUserIDFromContext(r.Context())
			if !ok || userID == "" {
				// Fall back to IP-based rate limiting for unauthenticated requests
				return httprate.KeyByIP(r)
			}
			return "user:" + userID, nil
		}),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			writeJSONError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests, please try again later")
		}),
	)
}

// StrictRateLimit creates a stricter rate limiting middleware for sensitive endpoints
func StrictRateLimit() func(http.Handler) http.Handler {
	return httprate.Limit(
		10,  // 10 requests
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			writeJSONError(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "Too many requests, please try again later")
		}),
	)
}
