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

# Order service specific
export PORT=8083
export DB_NAME=order_db
export SERVICE_NAME=order-service
export OUTBOX_POLL_INTERVAL=5s
export OUTBOX_BATCH_SIZE=100

# Run the service
echo "Starting Order Service on port 8083..."
echo "Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
./bin/order
