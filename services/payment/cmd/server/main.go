package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gocommerce/pkg/logger"
	"gocommerce/pkg/middleware"
	"gocommerce/pkg/tracing"
	api "gocommerce/services/payment/internal/adapters/http"
	"gocommerce/services/payment/internal/adapters/kafka"
	"gocommerce/services/payment/internal/adapters/postgres"
	paymentprovider "gocommerce/services/payment/internal/adapters/provider"
	"gocommerce/services/payment/internal/application"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
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

	log.Infof("Starting Payment Service on port %s", cfg.ServerPort)

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

	// Connect to database
	db, err := sql.Open("postgres", cfg.DBConnectionString())
	if err != nil {
		log.ErrorWithErr(err, "Failed to open database connection")
		os.Exit(1)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.ErrorWithErr(err, "Failed to ping database")
		os.Exit(1)
	}

	log.Info("Database connection established")

	// Initialize Kafka event publisher
	eventPublisher := kafka.NewEventPublisher(cfg.KafkaBrokers, log)
	defer eventPublisher.Close()

	log.Infof("Kafka publisher initialized with brokers: %v", cfg.KafkaBrokers)

	// Initialize payment provider (mock for now)
	paymentProvider := paymentprovider.NewMockPaymentProvider(log, cfg.ProviderFailureRate)

	log.Infof("Mock payment provider initialized with %.1f%% failure rate", cfg.ProviderFailureRate*100)

	// Initialize repository
	paymentRepo := postgres.NewPaymentRepository(db)

	// Initialize application service
	paymentService := application.NewPaymentService(paymentRepo, paymentProvider, eventPublisher)

	// Initialize HTTP handler
	handler := api.NewPaymentHandler(paymentService)

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

	log.Info("Server stopped")
}

// setupRouter configures the chi router with middleware and routes
func setupRouter(log *logger.Logger, handler *api.PaymentHandler) *chi.Mux {
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
