#!/bin/bash
set -e

# Script to create multiple databases in PostgreSQL
# Usage: Set POSTGRES_MULTIPLE_DATABASES environment variable with comma-separated database names

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE product_db;
    CREATE DATABASE order_db;
    CREATE DATABASE payment_db;
    CREATE DATABASE inventory_db;

    GRANT ALL PRIVILEGES ON DATABASE product_db TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE order_db TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE payment_db TO $POSTGRES_USER;
    GRANT ALL PRIVILEGES ON DATABASE inventory_db TO $POSTGRES_USER;
EOSQL

echo "Multiple databases created successfully!"
