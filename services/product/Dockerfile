# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.work go.work.sum ./
COPY pkg/go.mod pkg/go.sum ./pkg/
COPY services/product/go.mod services/product/go.sum ./services/product/

# Download dependencies
WORKDIR /build/services/product
RUN go mod download

# Copy source code
WORKDIR /build
COPY pkg/ ./pkg/
COPY services/product/ ./services/product/

# Build the application
WORKDIR /build/services/product
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a \
    -o /app/product-service \
    ./cmd/server

# Final stage
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary
COPY --from=builder /app/product-service /product-service

# Expose port
EXPOSE 8080

# Set user (non-root)
USER 65534:65534

# Run the application
ENTRYPOINT ["/product-service"]
