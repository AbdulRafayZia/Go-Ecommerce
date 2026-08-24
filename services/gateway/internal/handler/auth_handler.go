package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"gocommerce/pkg/logger"
	"gocommerce/services/gateway/internal/auth"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	jwtManager *auth.JWTManager
	userStore  auth.UserStore
	logger     *logger.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(jwtManager *auth.JWTManager, userStore auth.UserStore, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		jwtManager: jwtManager,
		userStore:  userStore,
		logger:     logger,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"`
	User         *UserInfo `json:"user"`
}

// UserInfo represents user information in the response
type UserInfo struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Login handles user login
// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("Invalid login request: %v", err)
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate credentials
	user, err := h.userStore.ValidateCredentials(req.Username, req.Password)
	if err != nil {
		h.logger.Warnf("Login failed for user %s: %v", req.Username, err)
		writeJSONError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid username or password")
		return
	}

	// Generate tokens
	tokenPair, err := h.jwtManager.GenerateTokenPair(user)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to generate tokens")
		writeJSONError(w, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate authentication tokens")
		return
	}

	// Build response
	response := LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenPair.ExpiresIn,
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Roles:    user.Roles,
		},
	}

	h.logger.Infof("User %s logged in successfully", user.Username)
	writeJSON(w, http.StatusOK, response)
}

// RefreshToken handles token refresh
// POST /auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warnf("Invalid refresh token request: %v", err)
		writeJSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate and refresh token
	newAccessToken, err := h.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		h.logger.Warnf("Token refresh failed: %v", err)
		writeJSONError(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Invalid or expired refresh token")
		return
	}

	// Get user info from refresh token
	claims, _ := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	user, err := h.userStore.GetByID(claims.UserID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get user info")
		writeJSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user information")
		return
	}

	// Build response
	response := LoginResponse{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600, // 1 hour
		User: &UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Roles:    user.Roles,
		},
	}

	h.logger.Infof("Token refreshed for user %s", user.Username)
	writeJSON(w, http.StatusOK, response)
}

// Logout handles user logout
// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// In a production system, you would:
	// 1. Add the token to a blacklist/revocation list
	// 2. Store in Redis with TTL matching token expiry
	// 3. Verify against blacklist in auth middleware

	// For now, just return success
	// The client is responsible for deleting the token
	h.logger.Info("User logged out")
	w.WriteHeader(http.StatusNoContent)
}

// HealthCheck returns health status
// GET /health
func (h *AuthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "api-gateway",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, response)
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Code:    code,
		Message: message,
	})
}
