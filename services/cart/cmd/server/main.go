package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gocommerce/pkg/logger"
	"gocommerce/pkg/middleware"
	"gocommerce/pkg/tracing"
	api "gocommerce/services/cart/internal/adapters/http"
	"gocommerce/services/cart/internal/adapters/redis"
	"gocommerce/services/cart/internal/application"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	goredis "github.com/redis/go-redis/v9"
)

func main() {
	// Load configuration
	cfg := LoadConfig()

	// Initialize logger
	log := logger.New(logger.Config{
		Level:       cfg.LogLevel,
		PrettyPrint: cfg.Environment == "development",
	}).WithFields(map[string]interface{}{
		"service": cfg.ServiceName,
		"env":     cfg.Environment,
	})

	log.Infof("Starting Cart Service on port %s", cfg.ServerPort)

	// Initialize tracing
	tracingCfg := tracing.Config{
		ServiceName:    cfg.ServiceName,
		ServiceVersion: "1.0.0",
		Environment:    cfg.Environment,
		OTLPEndpoint:   cfg.OTLPEndpoint,
		Enabled:        true,
	}

	provider, err := tracing.InitTracer(tracingCfg)
	if err != nil {
		log.ErrorWithErr(err, "Failed to initialize tracing")
		os.Exit(1)
	}
	defer provider.Shutdown(context.Background())

	// Connect to Redis
	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Verify Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.ErrorWithErr(err, "Failed to connect to Redis")
		os.Exit(1)
	}

	log.Info("Redis connection established")

	// Initialize repository
	cartRepo := redis.NewCartRepository(redisClient)

	// Initialize application service
	cartService := application.NewCartServiceWithTTL(cartRepo, cfg.CartTTL)

	// Initialize HTTP handler
	handler := api.NewCartHandler(cartService)

	// Setup router
	router := setupRouter(log, handler)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.ErrorWithErr(err, "Server failed to start")
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.ErrorWithErr(err, "Server forced to shutdown")
	}

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.ErrorWithErr(err, "Failed to close Redis connection")
	}

	log.Info("Server stopped")
}

// setupRouter configures the chi router with middleware and routes
func setupRouter(log *logger.Logger, handler *api.CartHandler) *chi.Mux {
	r := chi.NewRouter()

	// Apply middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.HTTPRecovery(log))
	r.Use(middleware.HTTPCorrelationID())
	r.Use(middleware.HTTPTracing())
	r.Use(middleware.HTTPLogging(log))
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// Register routes using the generated server
	api.HandlerFromMux(handler, r)

	return r
}
