package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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

	// Kafka configuration
	KafkaBrokers       []string
	KafkaConsumerGroup string
	KafkaTopics        []string

	// Reservation expiry configuration
	ReservationExpiryDuration time.Duration
	ReservationCleanupInterval time.Duration

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
		ServerPort:                 getEnv("PORT", "8083"),
		DBHost:                     getEnv("DB_HOST", "localhost"),
		DBPort:                     getEnv("DB_PORT", "5432"),
		DBUser:                     getEnv("DB_USER", "postgres"),
		DBPassword:                 getEnv("DB_PASSWORD", "postgres"),
		DBName:                     getEnv("DB_NAME", "inventory_db"),
		DBSSLMode:                  getEnv("DB_SSLMODE", "disable"),
		KafkaBrokers:               getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		KafkaConsumerGroup:         getEnv("KAFKA_CONSUMER_GROUP", "inventory-service"),
		KafkaTopics:                getEnvAsSlice("KAFKA_TOPICS", []string{"order.created", "order.paid", "order.cancelled"}),
		ReservationExpiryDuration:  getEnvAsDuration("RESERVATION_EXPIRY_DURATION", 30*time.Minute),
		ReservationCleanupInterval: getEnvAsDuration("RESERVATION_CLEANUP_INTERVAL", 5*time.Minute),
		OTLPEndpoint:               getEnv("OTLP_ENDPOINT", "localhost:4317"),
		ServiceName:                getEnv("SERVICE_NAME", "inventory-service"),
		Environment:                getEnv("ENVIRONMENT", "development"),
		LogLevel:                   getEnv("LOG_LEVEL", "info"),
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

// getEnvAsDuration gets an environment variable as a duration with a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsSlice gets an environment variable as a comma-separated slice
func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}
