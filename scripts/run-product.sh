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

# Product service specific
export PORT=8081
export DB_NAME=product_db
export SERVICE_NAME=product-service

# Run the service
echo "Starting Product Service on port 8081..."
echo "Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
./bin/product
