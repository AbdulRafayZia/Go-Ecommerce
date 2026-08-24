#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting all GoCommerce services...${NC}"
echo ""

# Check if binaries exist
if [ ! -f "./bin/product" ]; then
    echo -e "${YELLOW}Binaries not found. Building all services...${NC}"
    make build
fi

# Start services in background
echo -e "${GREEN}Starting Product Service...${NC}"
./scripts/run-product.sh > logs/product.log 2>&1 &
PRODUCT_PID=$!
echo "  PID: $PRODUCT_PID (logs: logs/product.log)"

sleep 2

echo -e "${GREEN}Starting Cart Service...${NC}"
./scripts/run-cart.sh > logs/cart.log 2>&1 &
CART_PID=$!
echo "  PID: $CART_PID (logs: logs/cart.log)"

sleep 2

echo -e "${GREEN}Starting Order Service...${NC}"
./scripts/run-order.sh > logs/order.log 2>&1 &
ORDER_PID=$!
echo "  PID: $ORDER_PID (logs: logs/order.log)"

sleep 2

echo -e "${GREEN}Starting Payment Service...${NC}"
./scripts/run-payment.sh > logs/payment.log 2>&1 &
PAYMENT_PID=$!
echo "  PID: $PAYMENT_PID (logs: logs/payment.log)"

sleep 2

echo -e "${GREEN}Starting Inventory Service...${NC}"
./scripts/run-inventory.sh > logs/inventory.log 2>&1 &
INVENTORY_PID=$!
echo "  PID: $INVENTORY_PID (logs: logs/inventory.log)"

sleep 2

echo -e "${GREEN}Starting API Gateway...${NC}"
./scripts/run-gateway.sh > logs/gateway.log 2>&1 &
GATEWAY_PID=$!
echo "  PID: $GATEWAY_PID (logs: logs/gateway.log)"

echo ""
echo -e "${GREEN}All services started!${NC}"
echo ""
echo "Service URLs:"
echo "  - API Gateway:  http://localhost:8000"
echo "  - Product:      http://localhost:8081"
echo "  - Cart:         http://localhost:8082"
echo "  - Order:        http://localhost:8083"
echo "  - Payment:      http://localhost:8084"
echo "  - Inventory:    http://localhost:8085"
echo ""
echo "PIDs: $GATEWAY_PID $PRODUCT_PID $CART_PID $ORDER_PID $PAYMENT_PID $INVENTORY_PID"
echo ""
echo "To stop all services, run: ./scripts/stop-all.sh"
echo "To view logs: tail -f logs/*.log"

# Save PIDs for stop script
echo "$GATEWAY_PID $PRODUCT_PID $CART_PID $ORDER_PID $PAYMENT_PID $INVENTORY_PID" > .service-pids
