#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${RED}Stopping all GoCommerce services...${NC}"

if [ -f ".service-pids" ]; then
    PIDS=$(cat .service-pids)
    for PID in $PIDS; do
        if ps -p $PID > /dev/null; then
            echo "Stopping process $PID..."
            kill $PID
        fi
    done
    rm .service-pids
    echo -e "${GREEN}All services stopped!${NC}"
else
    echo "No PID file found. Searching for running services..."
    pkill -f "bin/gateway"
    pkill -f "bin/product"
    pkill -f "bin/cart"
    pkill -f "bin/order"
    pkill -f "bin/payment"
    pkill -f "bin/inventory"
    echo -e "${GREEN}Services stopped!${NC}"
fi
