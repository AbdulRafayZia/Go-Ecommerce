#!/bin/bash

# Load common environment variables
export OTLP_ENDPOINT=localhost:4317
export ENVIRONMENT=development
export LOG_LEVEL=info

# Gateway specific
export SERVER_PORT=8000
export JWT_SECRET_KEY=your-secret-key-change-in-production
export ACCESS_TOKEN_TTL=1h
export REFRESH_TOKEN_TTL=168h
export RATE_LIMIT_ENABLED=true
export RATE_LIMIT_PER_MINUTE=100
export CORS_ENABLED=true

# Backend service URLs
export PRODUCT_SERVICE_URL=http://localhost:8081
export CART_SERVICE_URL=http://localhost:8082
export ORDER_SERVICE_URL=http://localhost:8083
export PAYMENT_SERVICE_URL=http://localhost:8084
export INVENTORY_SERVICE_URL=http://localhost:8085
export SERVICE_TIMEOUT=30s

# Run the service
echo "Starting API Gateway on port 8000..."
echo "Backend services:"
echo "  - Product:   $PRODUCT_SERVICE_URL"
echo "  - Cart:      $CART_SERVICE_URL"
echo "  - Order:     $ORDER_SERVICE_URL"
echo "  - Payment:   $PAYMENT_SERVICE_URL"
echo "  - Inventory: $INVENTORY_SERVICE_URL"
./bin/gateway
