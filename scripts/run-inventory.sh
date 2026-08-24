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

# Inventory service specific
export PORT=8085
export DB_NAME=inventory_db
export SERVICE_NAME=inventory-service
export KAFKA_GROUP_ID=inventory-service
export KAFKA_TOPICS=order.created,order.paid,order.cancelled
export RESERVATION_EXPIRY_DURATION=30m
export RESERVATION_CLEANUP_INTERVAL=5m

# Run the service
echo "Starting Inventory Service on port 8085..."
echo "Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
./bin/inventory
