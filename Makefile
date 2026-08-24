.PHONY: help proto test lint build clean dev down migrate-up migrate-down docker-build docker-push k8s-deploy

# Colors for terminal output
YELLOW := \033[1;33m
NC := \033[0m # No Color

help: ## Show this help message
	@echo '$(YELLOW)GoCommerce - Available Commands:$(NC)'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# Development
dev: ## Start all services with docker-compose
	docker compose -f deploy/docker-compose.yml up --build

down: ## Stop all services
	docker compose -f deploy/docker-compose.yml down

dev-detached: ## Start all services in detached mode
	docker compose -f deploy/docker-compose.yml up -d --build

logs: ## Show logs from all services
	docker compose -f deploy/docker-compose.yml logs -f

# API Code Generation
api-gen: ## Generate code from OpenAPI specifications
	@echo "Generating API code from OpenAPI specs..."
	@for service in product cart order payment inventory; do \
		if [ -f services/$$service/api/openapi.yaml ]; then \
			echo "Generating code for $$service..."; \
			$$(go env GOPATH)/bin/oapi-codegen -package api -generate types,chi-server \
				services/$$service/api/openapi.yaml > services/$$service/internal/adapters/http/generated.go || exit 1; \
		fi \
	done
	@echo "API code generated successfully!"

api-gen-product: ## Generate code for product service only
	@echo "Generating API code for product service..."
	@$$(go env GOPATH)/bin/oapi-codegen -package api -generate types,chi-server \
		services/product/api/openapi.yaml > services/product/internal/adapters/http/generated.go
	@echo "Product API code generated!"

api-lint: ## Lint OpenAPI specifications
	@echo "Linting OpenAPI specs..."
	@for spec in services/*/api/openapi.yaml; do \
		echo "Linting $$spec..."; \
		npx @redocly/cli lint "$$spec" || true; \
	done

# Testing
test: ## Run all tests
	@echo "Running tests..."
	go test ./... -v -race -coverprofile=coverage.out

test-product: ## Run product service tests only
	@echo "Running product service tests..."
	go test gocommerce/services/product/... -v -race

test-coverage: test ## Run tests and show coverage
	go tool cover -html=coverage.out

test-short: ## Run short tests only
	go test ./... -short -v

test-integration: ## Run integration tests
	go test ./... -tags=integration -v

# Linting
lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

lint-fix: ## Run linters and auto-fix issues
	golangci-lint run ./... --fix

# Build
build: ## Build all services
	@echo "Building all services..."
	@mkdir -p bin
	@for service in gateway product cart order payment inventory notification; do \
		if [ -f services/$$service/cmd/server/main.go ]; then \
			echo "Building $$service..."; \
			go build -o bin/$$service gocommerce/services/$$service/cmd/server; \
		fi \
	done
	@echo "Build complete!"

build-service: ## Build a specific service (usage: make build-service SERVICE=product)
	@echo "Building $(SERVICE)..."
	@mkdir -p bin
	@go build -o bin/$(SERVICE) gocommerce/services/$(SERVICE)/cmd/server

build-product: ## Build product service
	@echo "Building product service..."
	@mkdir -p bin
	@go build -o bin/product gocommerce/services/product/cmd/server

# Clean
clean: ## Clean build artifacts and generated files
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out
	go clean -cache
	@echo "Clean complete!"

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Docker
docker-build: ## Build Docker images for all services
	@echo "Building Docker images..."
	@for service in gateway product cart order payment inventory notification; do \
		echo "Building $$service image..."; \
		docker build -t gocommerce/$$service:latest -f services/$$service/Dockerfile .; \
	done
	@echo "Docker images built!"

docker-build-service: ## Build Docker image for a specific service (usage: make docker-build-service SERVICE=product)
	@echo "Building Docker image for $(SERVICE)..."
	docker build -t gocommerce/$(SERVICE):latest -f services/$(SERVICE)/Dockerfile .

docker-push: ## Push Docker images to registry
	@echo "Pushing Docker images..."
	@for service in gateway product cart order payment inventory notification; do \
		echo "Pushing $$service image..."; \
		docker push gocommerce/$$service:latest; \
	done

# Database Migrations
# Default database connection strings for local development
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= gocommerce
DB_PASSWORD ?= gocommerce_dev_password
DB_SSLMODE ?= disable

PRODUCT_DB ?= product_db
ORDER_DB ?= order_db
PAYMENT_DB ?= payment_db
INVENTORY_DB ?= inventory_db

# Connection string builders
PRODUCT_DB_URL = "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(PRODUCT_DB)?sslmode=$(DB_SSLMODE)"
ORDER_DB_URL = "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(ORDER_DB)?sslmode=$(DB_SSLMODE)"
PAYMENT_DB_URL = "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(PAYMENT_DB)?sslmode=$(DB_SSLMODE)"
INVENTORY_DB_URL = "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(INVENTORY_DB)?sslmode=$(DB_SSLMODE)"

# Create new migration
migrate-create: ## Create a new migration (usage: make migrate-create SERVICE=product NAME=create_products_table)
	@echo "Creating migration for $(SERVICE)..."
	@cd services/$(SERVICE)/migrations && goose create $(NAME) sql

# ==================== Product Service Migrations ====================
migrate-product-up: ## Run product service migrations up
	@echo "Running product service migrations up..."
	@cd services/product/migrations && goose postgres $(PRODUCT_DB_URL) up

migrate-product-down: ## Rollback product service migrations (1 step)
	@echo "Rolling back product service migrations..."
	@cd services/product/migrations && goose postgres $(PRODUCT_DB_URL) down

migrate-product-reset: ## Reset product service migrations (down all, then up)
	@echo "Resetting product service migrations..."
	@cd services/product/migrations && goose postgres $(PRODUCT_DB_URL) reset
	@cd services/product/migrations && goose postgres $(PRODUCT_DB_URL) up

migrate-product-status: ## Show product service migration status
	@echo "Product service migration status:"
	@cd services/product/migrations && goose postgres $(PRODUCT_DB_URL) status

# ==================== Order Service Migrations ====================
migrate-order-up: ## Run order service migrations up
	@echo "Running order service migrations up..."
	@cd services/order/migrations && goose postgres $(ORDER_DB_URL) up

migrate-order-down: ## Rollback order service migrations (1 step)
	@echo "Rolling back order service migrations..."
	@cd services/order/migrations && goose postgres $(ORDER_DB_URL) down

migrate-order-reset: ## Reset order service migrations (down all, then up)
	@echo "Resetting order service migrations..."
	@cd services/order/migrations && goose postgres $(ORDER_DB_URL) reset
	@cd services/order/migrations && goose postgres $(ORDER_DB_URL) up

migrate-order-status: ## Show order service migration status
	@echo "Order service migration status:"
	@cd services/order/migrations && goose postgres $(ORDER_DB_URL) status

# ==================== Payment Service Migrations ====================
migrate-payment-up: ## Run payment service migrations up
	@echo "Running payment service migrations up..."
	@cd services/payment/migrations && goose postgres $(PAYMENT_DB_URL) up

migrate-payment-down: ## Rollback payment service migrations (1 step)
	@echo "Rolling back payment service migrations..."
	@cd services/payment/migrations && goose postgres $(PAYMENT_DB_URL) down

migrate-payment-reset: ## Reset payment service migrations (down all, then up)
	@echo "Resetting payment service migrations..."
	@cd services/payment/migrations && goose postgres $(PAYMENT_DB_URL) reset
	@cd services/payment/migrations && goose postgres $(PAYMENT_DB_URL) up

migrate-payment-status: ## Show payment service migration status
	@echo "Payment service migration status:"
	@cd services/payment/migrations && goose postgres $(PAYMENT_DB_URL) status

# ==================== Inventory Service Migrations ====================
migrate-inventory-up: ## Run inventory service migrations up
	@echo "Running inventory service migrations up..."
	@cd services/inventory/migrations && goose postgres $(INVENTORY_DB_URL) up

migrate-inventory-down: ## Rollback inventory service migrations (1 step)
	@echo "Rolling back inventory service migrations..."
	@cd services/inventory/migrations && goose postgres $(INVENTORY_DB_URL) down

migrate-inventory-reset: ## Reset inventory service migrations (down all, then up)
	@echo "Resetting inventory service migrations..."
	@cd services/inventory/migrations && goose postgres $(INVENTORY_DB_URL) reset
	@cd services/inventory/migrations && goose postgres $(INVENTORY_DB_URL) up

migrate-inventory-status: ## Show inventory service migration status
	@echo "Inventory service migration status:"
	@cd services/inventory/migrations && goose postgres $(INVENTORY_DB_URL) status

# ==================== All Services Migrations ====================
migrate-up-all: ## Run migrations for all services (product, order, payment, inventory)
	@echo "Running migrations for all services..."
	@$(MAKE) migrate-product-up
	@$(MAKE) migrate-order-up
	@$(MAKE) migrate-payment-up
	@$(MAKE) migrate-inventory-up
	@echo "All migrations complete!"

migrate-down-all: ## Rollback migrations for all services (1 step each)
	@echo "Rolling back migrations for all services..."
	@$(MAKE) migrate-inventory-down
	@$(MAKE) migrate-payment-down
	@$(MAKE) migrate-order-down
	@$(MAKE) migrate-product-down
	@echo "All rollbacks complete!"

migrate-reset-all: ## Reset migrations for all services (WARNING: destroys all data)
	@echo "⚠️  WARNING: This will destroy all data in all databases!"
	@echo "Resetting migrations for all services..."
	@$(MAKE) migrate-product-reset
	@$(MAKE) migrate-order-reset
	@$(MAKE) migrate-payment-reset
	@$(MAKE) migrate-inventory-reset
	@echo "All migrations reset!"

migrate-status-all: ## Show migration status for all services
	@echo "=== Migration Status for All Services ==="
	@echo ""
	@$(MAKE) migrate-product-status
	@echo ""
	@$(MAKE) migrate-order-status
	@echo ""
	@$(MAKE) migrate-payment-status
	@echo ""
	@$(MAKE) migrate-inventory-status
	@echo ""
	@echo "=== End of Migration Status ==="

# Kubernetes
k8s-deploy: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	kubectl apply -f deploy/k8s/base/

k8s-delete: ## Delete from Kubernetes
	@echo "Deleting from Kubernetes..."
	kubectl delete -f deploy/k8s/base/

k8s-logs: ## Show logs from a service in Kubernetes (usage: make k8s-logs SERVICE=product)
	kubectl logs -f deployment/$(SERVICE)

# Code Quality
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	gofmt -s -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	godoc -http=:6060

# Install tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "Tools installed!"

# Quick start
init: install-tools deps ## Initialize the project
	@echo "Project initialized! Run 'make dev' to start."

# Run all quality checks
check: lint test ## Run all quality checks

# Full CI pipeline
ci: lint test build ## Run full CI pipeline
	@echo "CI pipeline complete!"
