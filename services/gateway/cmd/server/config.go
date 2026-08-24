package main

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	ServerPort string
	Environment string

	// JWT configuration
	JWTSecretKey       string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration

	// Rate limiting
	RateLimitEnabled      bool
	RateLimitPerMinute    int
	RateLimitBurstSize    int

	// Backend service URLs
	ProductServiceURL   string
	CartServiceURL      string
	OrderServiceURL     string
	PaymentServiceURL   string
	InventoryServiceURL string

	// Timeouts
	ServiceTimeout time.Duration

	// Observability
	OTLPEndpoint string
	LogLevel     string

	// CORS
	CORSEnabled        bool
	CORSAllowedOrigins []string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		// Server
		ServerPort:  getEnv("SERVER_PORT", "8000"),
		Environment: getEnv("ENVIRONMENT", "development"),

		// JWT
		JWTSecretKey:    getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
		AccessTokenTTL:  parseDuration(getEnv("ACCESS_TOKEN_TTL", "1h")),
		RefreshTokenTTL: parseDuration(getEnv("REFRESH_TOKEN_TTL", "168h")), // 7 days

		// Rate limiting
		RateLimitEnabled:   parseBool(getEnv("RATE_LIMIT_ENABLED", "true")),
		RateLimitPerMinute: parseInt(getEnv("RATE_LIMIT_PER_MINUTE", "100")),
		RateLimitBurstSize: parseInt(getEnv("RATE_LIMIT_BURST_SIZE", "20")),

		// Backend services
		ProductServiceURL:   getEnv("PRODUCT_SERVICE_URL", "http://localhost:8081"),
		CartServiceURL:      getEnv("CART_SERVICE_URL", "http://localhost:8082"),
		OrderServiceURL:     getEnv("ORDER_SERVICE_URL", "http://localhost:8083"),
		PaymentServiceURL:   getEnv("PAYMENT_SERVICE_URL", "http://localhost:8084"),
		InventoryServiceURL: getEnv("INVENTORY_SERVICE_URL", "http://localhost:8085"),

		// Timeouts
		ServiceTimeout: parseDuration(getEnv("SERVICE_TIMEOUT", "30s")),

		// Observability
		OTLPEndpoint: getEnv("OTLP_ENDPOINT", "localhost:4317"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),

		// CORS
		CORSEnabled:        parseBool(getEnv("CORS_ENABLED", "true")),
		CORSAllowedOrigins: []string{"http://localhost:3000", "http://localhost:8000"},
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseDuration parses a duration string
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}

// parseInt parses an integer string
func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// parseBool parses a boolean string
func parseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}
