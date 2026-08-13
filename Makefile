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
migrate-create: ## Create a new migration (usage: make migrate-create SERVICE=order NAME=create_orders_table)
	@echo "Creating migration for $(SERVICE)..."
	migrate create -ext sql -dir services/$(SERVICE)/migrations -seq $(NAME)

migrate-up: ## Run migrations up for a service (usage: make migrate-up SERVICE=order)
	@echo "Running migrations up for $(SERVICE)..."
	migrate -path services/$(SERVICE)/migrations -database "$(DB_URL)" up

migrate-down: ## Run migrations down for a service (usage: make migrate-down SERVICE=order)
	@echo "Running migrations down for $(SERVICE)..."
	migrate -path services/$(SERVICE)/migrations -database "$(DB_URL)" down

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
