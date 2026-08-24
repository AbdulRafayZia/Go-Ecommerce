package handler

import (
	"net/http"
	"strings"
	"time"

	"gocommerce/pkg/logger"
	"gocommerce/services/gateway/internal/proxy"
)

// GatewayHandler handles gateway routing and proxying
type GatewayHandler struct {
	proxy  *proxy.ReverseProxy
	logger *logger.Logger
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(proxy *proxy.ReverseProxy, logger *logger.Logger) *GatewayHandler {
	return &GatewayHandler{
		proxy:  proxy,
		logger: logger,
	}
}

// RouteRequest routes incoming requests to the appropriate backend service
func (h *GatewayHandler) RouteRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Determine service and target path based on URL prefix
	var serviceName, targetPath string

	switch {
	case strings.HasPrefix(path, "/api/products"):
		serviceName = "product"
		targetPath = strings.TrimPrefix(path, "/api")

	case strings.HasPrefix(path, "/api/carts"):
		serviceName = "cart"
		targetPath = strings.TrimPrefix(path, "/api")

	case strings.HasPrefix(path, "/api/orders"):
		serviceName = "order"
		targetPath = strings.TrimPrefix(path, "/api")

	case strings.HasPrefix(path, "/api/payments"):
		serviceName = "payment"
		targetPath = strings.TrimPrefix(path, "/api")

	case strings.HasPrefix(path, "/api/inventory"):
		serviceName = "inventory"
		targetPath = strings.TrimPrefix(path, "/api")

	default:
		h.logger.Warnf("No service route found for path: %s", path)
		writeJSONError(w, http.StatusNotFound, "ROUTE_NOT_FOUND", "No service route found for this path")
		return
	}

	// Log routing decision
	h.logger.Infof("Routing %s %s -> %s%s", r.Method, path, serviceName, targetPath)

	// Proxy the request
	h.proxy.ProxyRequest(w, r, serviceName, targetPath)
}

// HealthCheck performs health checks on all backend services and returns aggregated status
func (h *GatewayHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type ServiceHealth struct {
		Status    string `json:"status"`
		LatencyMs int64  `json:"latency_ms,omitempty"`
		LastCheck string `json:"last_check"`
	}

	type HealthResponse struct {
		Status       string                    `json:"status"`
		Service      string                    `json:"service"`
		Timestamp    string                    `json:"timestamp"`
		Dependencies map[string]ServiceHealth  `json:"dependencies,omitempty"`
	}

	// Check all backend services
	services := []string{"product", "cart", "order", "payment", "inventory"}
	dependencies := make(map[string]ServiceHealth)
	overallHealthy := true

	for _, svc := range services {
		start := time.Now()
		err := h.proxy.CheckHealth(svc)
		latency := time.Since(start).Milliseconds()

		status := "healthy"
		if err != nil {
			status = "unhealthy"
			overallHealthy = false
			h.logger.Warnf("Service %s is unhealthy: %v", svc, err)
		}

		dependencies[svc] = ServiceHealth{
			Status:    status,
			LatencyMs: latency,
			LastCheck: time.Now().Format(time.RFC3339),
		}
	}

	overallStatus := "healthy"
	if !overallHealthy {
		overallStatus = "degraded"
	}

	response := HealthResponse{
		Status:       overallStatus,
		Service:      "api-gateway",
		Timestamp:    time.Now().Format(time.RFC3339),
		Dependencies: dependencies,
	}

	// Return 503 if any service is down
	statusCode := http.StatusOK
	if !overallHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, response)
}
