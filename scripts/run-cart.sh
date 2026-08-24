#!/bin/bash

# Load common environment variables
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
export REDIS_DB=0
export OTLP_ENDPOINT=localhost:4317
export ENVIRONMENT=development
export LOG_LEVEL=info

# Cart service specific (uses Redis, not PostgreSQL)
export PORT=8083
export SERVICE_NAME=cart-service
export CART_TTL=168h

# Run the service
echo "Starting Cart Service on port 8083..."
echo "Redis: $REDIS_HOST:$REDIS_PORT"
./bin/cart
