package httputil

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// ClientConfig holds HTTP client configuration
type ClientConfig struct {
	BaseURL          string
	Timeout          time.Duration
	MaxIdleConns     int
	IdleConnTimeout  time.Duration
	EnableTracing    bool
	EnableRetry      bool
	RetryAttempts    int
	RetryWaitTime    time.Duration
}

// DefaultClientConfig returns default client configuration
func DefaultClientConfig(baseURL string) ClientConfig {
	return ClientConfig{
		BaseURL:         baseURL,
		Timeout:         30 * time.Second,
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
		EnableTracing:   true,
		EnableRetry:     false,
		RetryAttempts:   3,
		RetryWaitTime:   1 * time.Second,
	}
}

// NewClient creates a new HTTP client
func NewClient(cfg ClientConfig) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConns,
		IdleConnTimeout:     cfg.IdleConnTimeout,
	}

	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}

	// Add OpenTelemetry tracing if enabled
	if cfg.EnableTracing {
		client.Transport = otelhttp.NewTransport(transport)
	}

	return client
}

// DoWithRetry performs an HTTP request with retry logic
func DoWithRetry(ctx context.Context, client *http.Client, req *http.Request, attempts int, waitTime time.Duration) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i < attempts; i++ {
		resp, err = client.Do(req.WithContext(ctx))

		// If successful or non-retryable error, return
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		// Don't retry on last attempt
		if i < attempts-1 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
				// Exponential backoff
				waitTime *= 2
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", attempts, err)
	}

	return resp, nil
}
