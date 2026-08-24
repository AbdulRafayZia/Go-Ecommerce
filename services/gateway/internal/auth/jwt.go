package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrMissingClaims    = errors.New("missing required claims")
)

// Claims represents the JWT claims
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// TokenPair represents an access token and refresh token
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey        []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	issuer           string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, accessTokenTTL, refreshTokenTTL time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		issuer:          "gocommerce-gateway",
	}
}

// GenerateTokenPair generates an access token and refresh token for a user
func (m *JWTManager) GenerateTokenPair(user *User) (*TokenPair, error) {
	// Generate access token
	accessToken, err := m.generateToken(user, m.accessTokenTTL, "access")
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := m.generateToken(user, m.refreshTokenTTL, "refresh")
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(m.accessTokenTTL.Seconds()),
	}, nil
}

// generateToken generates a JWT token with the given TTL and token type
func (m *JWTManager) generateToken(user *User, ttl time.Duration, tokenType string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			ID:        uuid.New().String(),
			Audience:  jwt.ClaimStrings{tokenType},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, ErrInvalidSignature
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate required claims
	if claims.UserID == "" || claims.Username == "" {
		return nil, ErrMissingClaims
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Verify this is a refresh token
	if len(claims.Audience) == 0 || claims.Audience[0] != "refresh" {
		return nil, errors.New("not a refresh token")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a refresh token
func (m *JWTManager) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := m.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Create user from claims
	user := &User{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Roles:    claims.Roles,
	}

	// Generate new access token
	return m.generateToken(user, m.accessTokenTTL, "access")
}

// HasRole checks if the claims contain a specific role
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user is an admin
func (c *Claims) IsAdmin() bool {
	return c.HasRole("admin")
}
