package httputil

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gocommerce/pkg/logger"
	"gocommerce/pkg/middleware"

	"github.com/go-chi/chi/v5"
)

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port            int
	Logger          *logger.Logger
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port:            8080,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

// NewRouter creates a new chi router with common middleware
func NewRouter(cfg ServerConfig) *chi.Mux {
	r := chi.NewRouter()

	// Apply standard middleware
	r.Use(middleware.HTTPRecovery(cfg.Logger))
	r.Use(middleware.HTTPCorrelationID())
	r.Use(middleware.HTTPTracing())
	r.Use(middleware.HTTPLogging(cfg.Logger))

	return r
}

// RunServer runs the HTTP server with graceful shutdown
func RunServer(cfg ServerConfig, handler http.Handler) error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	cfg.Logger.Infof("Starting HTTP server on port %d", cfg.Port)

	// Channel to listen for errors from the server
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to serve: %w", err)
		}
	}()

	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either an error or an interrupt signal
	select {
	case err := <-errChan:
		return err
	case sig := <-quit:
		cfg.Logger.Infof("Received signal %v, initiating graceful shutdown", sig)
	}

	// Graceful shutdown
	return GracefulShutdown(cfg, server)
}

// GracefulShutdown performs graceful shutdown of the HTTP server
func GracefulShutdown(cfg ServerConfig, server *http.Server) error {
	cfg.Logger.Info("Shutting down HTTP server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		cfg.Logger.Warn("Shutdown timeout exceeded, forcing close")
		return server.Close()
	}

	cfg.Logger.Info("HTTP server stopped gracefully")
	return nil
}
