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


type routeMapping struct {
	gatewayPrefix string 
	serviceName   string 
	servicePrefix string 
	description   string 
}


func getRouteTable() []routeMapping {
	return []routeMapping{

		{
			gatewayPrefix: "/api/products",
			serviceName:   "product",
			servicePrefix: "/v1/products",
			description:   "Product catalog (list, get, create, update, delete products)",
		},
		{
			gatewayPrefix: "/api/categories",
			serviceName:   "product",
			servicePrefix: "/v1/categories",
			description:   "Product categories (list, create categories)",
		},
		{
			gatewayPrefix: "/api/carts",
			serviceName:   "cart",
			servicePrefix: "/carts",
			description:   "Shopping carts (get, add items, update, remove items)",
		},

		// Order Service - handles order management
		{
			gatewayPrefix: "/api/orders",
			serviceName:   "order",
			servicePrefix: "/orders",
			description:   "Orders (list, create, get, update status, cancel)",
		},

		// Payment Service - handles payment processing
		{
			gatewayPrefix: "/api/payments",
			serviceName:   "payment",
			servicePrefix: "/payments",
			description:   "Payments (create, get, capture, cancel, refund)",
		},

		// Inventory Service - handles stock management
		{
			gatewayPrefix: "/api/inventory/stocks",
			serviceName:   "inventory",
			servicePrefix: "/stocks",
			description:   "Inventory stocks (get, add, set stock levels)",
		},
		{
			gatewayPrefix: "/api/inventory/reservations",
			serviceName:   "inventory",
			servicePrefix: "/reservations",
			description:   "Inventory reservations (get reservations by order)",
		},
	}
}

// RouteRequest routes incoming requests to the appropriate backend service
func (h *GatewayHandler) RouteRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path


	for _, route := range getRouteTable() {
		if strings.HasPrefix(path, route.gatewayPrefix) {

			remainingPath := strings.TrimPrefix(path, route.gatewayPrefix)

			targetPath := route.servicePrefix + remainingPath

			h.logger.Infof("Routing %s %s -> %s (%s)",
				r.Method,
				path,
				route.serviceName+":"+targetPath,
				route.description,
			)

			h.proxy.ProxyRequest(w, r, route.serviceName, targetPath)
			return
		}
	}
	h.logger.Warnf("No service route found for path: %s", path)
	writeJSONError(w, http.StatusNotFound, "ROUTE_NOT_FOUND", "No service route found for this path")
}

// HealthCheck performs health checks on all backend services and returns aggregated status
func (h *GatewayHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type ServiceHealth struct {
		Status    string `json:"status"`
		LatencyMs int64  `json:"latency_ms,omitempty"`
		LastCheck string `json:"last_check"`
	}

	type HealthResponse struct {
		Status       string                   `json:"status"`
		Service      string                   `json:"service"`
		Timestamp    string                   `json:"timestamp"`
		Dependencies map[string]ServiceHealth `json:"dependencies,omitempty"`
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
