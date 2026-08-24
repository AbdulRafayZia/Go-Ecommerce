package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gocommerce/pkg/logger"
)

// ServiceConfig holds configuration for a backend service
type ServiceConfig struct {
	Name    string
	BaseURL string
	Timeout time.Duration
	Healthy bool
}

// ReverseProxy handles proxying requests to backend services
type ReverseProxy struct {
	services map[string]*ServiceConfig
	client   *http.Client
	logger   *logger.Logger
}

// NewReverseProxy creates a new reverse proxy
func NewReverseProxy(logger *logger.Logger) *ReverseProxy {
	return &ReverseProxy{
		services: make(map[string]*ServiceConfig),
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		logger: logger,
	}
}

// RegisterService registers a backend service
func (p *ReverseProxy) RegisterService(name, baseURL string, timeout time.Duration) error {
	// Validate URL
	_, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid service URL: %w", err)
	}

	p.services[name] = &ServiceConfig{
		Name:    name,
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		Timeout: timeout,
		Healthy: true,
	}

	p.logger.Infof("Registered service: %s -> %s", name, baseURL)
	return nil
}

// ProxyRequest proxies a request to a backend service
func (p *ReverseProxy) ProxyRequest(w http.ResponseWriter, r *http.Request, serviceName, targetPath string) {
	p.logger.Infof("Service to check: %s, targetPath: %s", serviceName, targetPath)
	service, exists := p.services[serviceName]
	if !exists {
		p.logger.Errorf("Service not found: %s", serviceName)
		http.Error(w, "Service not found", http.StatusBadGateway)
		return
	}

	// Check service health
	if !service.Healthy {
		p.logger.Warnf("Service unhealthy: %s", serviceName)
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Build target URL
	targetURL := service.BaseURL + targetPath
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Create new request
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		p.logger.ErrorWithErr(err, "Failed to create proxy request")
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	p.copyHeaders(proxyReq.Header, r.Header)

	// Add/update headers
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
	proxyReq.Header.Set("X-Forwarded-Proto", r.URL.Scheme)
	proxyReq.Header.Set("X-Forwarded-Host", r.Host)

	// Set timeout
	ctx, cancel := context.WithTimeout(r.Context(), service.Timeout)
	defer cancel()
	proxyReq = proxyReq.WithContext(ctx)

	// Log request
	p.logger.Infof("Proxying %s %s -> %s", r.Method, r.URL.Path, targetURL)

	// Execute request
	start := time.Now()
	resp, err := p.client.Do(proxyReq)
	duration := time.Since(start)

	if err != nil {
		p.logger.ErrorWithErr(err, fmt.Sprintf("Proxy request failed for %s", serviceName))

		// Check if context was cancelled (timeout)
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "Gateway timeout", http.StatusGatewayTimeout)
		} else {
			http.Error(w, "Bad gateway", http.StatusBadGateway)
		}
		return
	}
	defer resp.Body.Close()

	// Log response
	p.logger.Infof("Proxy response from %s: %d (took %v)", serviceName, resp.StatusCode, duration)

	// Copy response headers
	p.copyHeaders(w.Header(), resp.Header)

	// Write status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		p.logger.ErrorWithErr(err, "Failed to copy response body")
	}
}

// copyHeaders copies HTTP headers from src to dst
func (p *ReverseProxy) copyHeaders(dst, src http.Header) {
	for key, values := range src {
		// Skip hop-by-hop headers
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// isHopByHopHeader checks if a header is hop-by-hop
func isHopByHopHeader(header string) bool {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	header = http.CanonicalHeaderKey(header)
	for _, h := range hopByHopHeaders {
		if header == h {
			return true
		}
	}
	return false
}

// CheckHealth performs a health check on a service
func (p *ReverseProxy) CheckHealth(serviceName string) error {
	service, exists := p.services[serviceName]
	if !exists {
		return fmt.Errorf("service not found: %s", serviceName)
	}

	// Try to hit the health endpoint
	healthURL := service.BaseURL + "/health"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		service.Healthy = false
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		service.Healthy = false
		return fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
	}

	service.Healthy = true
	return nil
}

// GetServiceHealth returns the health status of a service
func (p *ReverseProxy) GetServiceHealth(serviceName string) (bool, error) {
	service, exists := p.services[serviceName]
	if !exists {
		return false, fmt.Errorf("service not found: %s", serviceName)
	}

	return service.Healthy, nil
}

// GetAllServicesHealth returns health status of all services
func (p *ReverseProxy) GetAllServicesHealth() map[string]bool {
	health := make(map[string]bool)
	for name, service := range p.services {
		health[name] = service.Healthy
	}
	return health
}
