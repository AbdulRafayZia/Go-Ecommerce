package middleware

import (
	"context"
	"net/http"
	"strings"

	"gocommerce/pkg/logger"
	"gocommerce/services/gateway/internal/auth"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// ClaimsContextKey is the context key for JWT claims
	ClaimsContextKey contextKey = "jwt_claims"
	// UserIDContextKey is the context key for user ID
	UserIDContextKey contextKey = "user_id"
)

// AuthMiddleware creates a middleware that validates JWT tokens
func AuthMiddleware(jwtManager *auth.JWTManager, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Warn("Missing Authorization header")
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization token")
				return
			}

			// Check for Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Warn("Invalid Authorization header format")
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization format")
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				log.Warnf("Token validation failed: %v", err)
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)

			// Add user ID to logger context
			log = log.WithFields(map[string]interface{}{
				"user_id":  claims.UserID,
				"username": claims.Username,
			})

			// Continue with authenticated request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't fail if token is missing
func OptionalAuthMiddleware(jwtManager *auth.JWTManager, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No token provided, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				// Invalid format, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			tokenString := parts[1]
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				// Invalid token, continue without authentication
				log.Warnf("Optional token validation failed: %v", err)
				next.ServeHTTP(w, r)
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
			ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole creates a middleware that checks if the user has a specific role
func RequireRole(role string, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context
			claims, ok := r.Context().Value(ClaimsContextKey).(*auth.Claims)
			if !ok || claims == nil {
				log.Warn("No claims found in context for role check")
				writeJSONError(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
				return
			}

			// Check if user has the required role
			if !claims.HasRole(role) {
				log.Warnf("User %s missing required role: %s", claims.Username, role)
				writeJSONError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is a convenience middleware that requires admin role
func RequireAdmin(log *logger.Logger) func(http.Handler) http.Handler {
	return RequireRole("admin", log)
}

// GetClaimsFromContext retrieves JWT claims from the request context
func GetClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*auth.Claims)
	return claims, ok
}

// GetUserIDFromContext retrieves user ID from the request context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(string)
	return userID, ok
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"code":"` + code + `","message":"` + message + `"}`))
}
