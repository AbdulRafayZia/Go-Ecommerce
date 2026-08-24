#!/bin/bash

# Health check script for all GoCommerce services
# Usage: ./scripts/health-check.sh

echo "=========================================="
echo "GoCommerce Services Health Check"
echo "=========================================="
echo ""

# ANSI color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check service health
check_service() {
    local name=$1
    local port=$2
    local url="http://localhost:$port/health"

    printf "%-20s (:%s) ... " "$name" "$port"

    # Try to connect
    response=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 2 "$url" 2>/dev/null)

    if [ "$response" = "200" ]; then
        echo -e "${GREEN}✓ Healthy${NC}"
        return 0
    elif [ "$response" = "000" ]; then
        echo -e "${RED}✗ Not Running${NC}"
        return 1
    else
        echo -e "${YELLOW}⚠ Degraded (HTTP $response)${NC}"
        return 1
    fi
}

# Check all services
services_ok=0
services_total=6

echo "Microservices:"
echo "----------------------------------------"
check_service "API Gateway" "8000" && ((services_ok++))
check_service "Product Service" "8081" && ((services_ok++))
check_service "Order Service" "8082" && ((services_ok++))
check_service "Cart Service" "8083" && ((services_ok++))
check_service "Payment Service" "8084" && ((services_ok++))
check_service "Inventory Service" "8085" && ((services_ok++))

echo ""
echo "Infrastructure:"
echo "----------------------------------------"

# Check PostgreSQL
printf "%-20s ... " "PostgreSQL"
if nc -z localhost 5432 2>/dev/null; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not Running${NC}"
fi

# Check Redis
printf "%-20s ... " "Redis"
if nc -z localhost 6379 2>/dev/null; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not Running${NC}"
fi

# Check Kafka
printf "%-20s ... " "Kafka"
if nc -z localhost 9094 2>/dev/null; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not Running${NC}"
fi

echo ""
echo "=========================================="
echo "Summary: $services_ok/$services_total services healthy"
echo "=========================================="

if [ $services_ok -eq $services_total ]; then
    echo -e "${GREEN}All services are running!${NC}"
    exit 0
else
    echo -e "${YELLOW}Some services are not running.${NC}"
    echo ""
    echo "To start all services:"
    echo "  ./scripts/run-all.sh"
    echo ""
    echo "To check individual service logs:"
    echo "  tail -f /tmp/<service>.log"
    exit 1
fi
