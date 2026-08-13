package main

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	ServerPort string

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Observability
	OTLPEndpoint string
	ServiceName  string
	Environment  string

	// Logging
	LogLevel string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		ServerPort:   getEnv("PORT", "8080"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", "postgres"),
		DBName:       getEnv("DB_NAME", "product_db"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		OTLPEndpoint: getEnv("OTLP_ENDPOINT", "localhost:4317"),
		ServiceName:  getEnv("SERVICE_NAME", "product-service"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
	}
}

// DBConnectionString returns the PostgreSQL connection string
func (c *Config) DBConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
