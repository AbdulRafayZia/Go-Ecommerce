#!/bin/bash

# Load common environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=gocommerce
export DB_PASSWORD=gocommerce_dev_password
export DB_SSLMODE=disable
export KAFKA_BROKERS=localhost:9092
export OTLP_ENDPOINT=localhost:4317
export ENVIRONMENT=development
export LOG_LEVEL=info

# Payment service specific
export PORT=8084
export DB_NAME=payment_db
export SERVICE_NAME=payment-service
export PROVIDER_FAILURE_RATE=0.05

# Run the service
echo "Starting Payment Service on port 8084..."
echo "Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
./bin/payment
