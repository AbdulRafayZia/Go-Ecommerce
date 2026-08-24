#!/bin/bash

# Load common environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=gocommerce
export DB_PASSWORD=gocommerce_dev_password
export DB_SSLMODE=disable
export KAFKA_BROKERS=localhost:9094
export OTLP_ENDPOINT=localhost:4317
export ENVIRONMENT=development
export LOG_LEVEL=info

# Gateway specific
export PORT=8000
export SERVICE_NAME=api-gateway
export JWT_SECRET=your-secret-key-change-in-production-minimum-32-characters-long
export JWT_ISSUER=gocommerce-api-gateway
export JWT_ACCESS_TOKEN_TTL=1h
export JWT_REFRESH_TOKEN_TTL=168h

# Service URLs
export PRODUCT_SERVICE_URL=http://localhost:8081
export CART_SERVICE_URL=http://localhost:8083
export ORDER_SERVICE_URL=http://localhost:8082
export PAYMENT_SERVICE_URL=http://localhost:8084
export INVENTORY_SERVICE_URL=http://localhost:8085

# Run the service
echo "Starting API Gateway on port 8000..."
echo "Product Service: $PRODUCT_SERVICE_URL"
echo "Cart Service: $CART_SERVICE_URL"
echo "Order Service: $ORDER_SERVICE_URL"
echo "Payment Service: $PAYMENT_SERVICE_URL"
echo "Inventory Service: $INVENTORY_SERVICE_URL"
./bin/gateway
