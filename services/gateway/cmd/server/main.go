package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"gocommerce/pkg/logger"
	pkgmiddleware "gocommerce/pkg/middleware"
	"gocommerce/pkg/tracing"
	"gocommerce/services/gateway/internal/auth"
	"gocommerce/services/gateway/internal/handler"
	"gocommerce/services/gateway/internal/middleware"
	"gocommerce/services/gateway/internal/proxy"
)

func main() {
	// Load configuration
	cfg := LoadConfig()

	// Initialize logger
	log := logger.New(logger.Config{
		Level:       cfg.LogLevel,
		PrettyPrint: cfg.Environment == "development",
	}).WithFields(map[string]interface{}{
		"service": "api-gateway",
		"env":     cfg.Environment,
	})

	log.Infof("Starting API Gateway on port %s", cfg.ServerPort)

	// Initialize tracing
	tracingCfg := tracing.Config{
		ServiceName:    "api-gateway",
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

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.JWTSecretKey,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	log.Info("JWT manager initialized")

	// Initialize user store (in-memory for demo)
	userStore := auth.NewInMemoryUserStore()
	log.Info("User store initialized with default users")

	// Initialize reverse proxy
	reverseProxy := proxy.NewReverseProxy(log)

	// Register backend services
	services := map[string]string{
		"product":   cfg.ProductServiceURL,
		"cart":      cfg.CartServiceURL,
		"order":     cfg.OrderServiceURL,
		"payment":   cfg.PaymentServiceURL,
		"inventory": cfg.InventoryServiceURL,
	}

	for name, url := range services {
		if err := reverseProxy.RegisterService(name, url, cfg.ServiceTimeout); err != nil {
			log.ErrorWithErr(err, "Failed to register service: "+name)
			os.Exit(1)
		}
	}

	// Initialize handlers
	authHandler := handler.NewAuthHandler(jwtManager, userStore, log)
	gatewayHandler := handler.NewGatewayHandler(reverseProxy, log)

	// Setup router
	router := setupRouter(cfg, log, jwtManager, authHandler, gatewayHandler)

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
		log.Infof("API Gateway listening on %s", server.Addr)
		log.Info("Available routes:")
		log.Info("  Authentication:")
		log.Info("    - POST /auth/login (login)")
		log.Info("    - POST /auth/refresh (refresh token)")
		log.Info("    - POST /auth/logout (logout, requires auth)")
		log.Info("    - GET  /auth/me (get current user, requires auth)")
		log.Info("  Health:")
		log.Info("    - GET  /health (gateway health check)")
		log.Info("  Public APIs:")
		log.Info("    - GET  /api/products (list products)")
		log.Info("    - GET  /api/products/{id} (get product)")
		log.Info("    - GET  /api/categories (list categories)")
		log.Info("    - GET  /api/categories/{id} (get category)")
		log.Info("  Protected APIs (require authentication):")
		log.Info("    - /api/products/* (create/update/delete - admin only)")
		log.Info("    - /api/categories/* (create/update/delete - admin only)")
		log.Info("    - /api/carts/* (cart management)")
		log.Info("    - /api/orders/* (order management)")
		log.Info("    - /api/payments/* (payment processing)")
		log.Info("    - /api/inventory/* (inventory management - admin only)")

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
func setupRouter(
	cfg *Config,
	log *logger.Logger,
	jwtManager *auth.JWTManager,
	authHandler *handler.AuthHandler,
	gatewayHandler *handler.GatewayHandler,
) *chi.Mux {
	r := chi.NewRouter()

	// Base middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(pkgmiddleware.HTTPRecovery(log))
	r.Use(pkgmiddleware.HTTPCorrelationID())
	r.Use(pkgmiddleware.HTTPTracing())
	r.Use(pkgmiddleware.HTTPLogging(log))

	// CORS middleware
	if cfg.CORSEnabled {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   cfg.CORSAllowedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	// Rate limiting (global)
	if cfg.RateLimitEnabled {
		rateLimitCfg := middleware.RateLimitConfig{
			RequestsPerMinute: cfg.RateLimitPerMinute,
			BurstSize:         cfg.RateLimitBurstSize,
		}
		r.Use(middleware.RateLimit(rateLimitCfg))
	}

	// Health check endpoint (no auth required)
	r.Get("/health", gatewayHandler.HealthCheck)

	// Authentication routes (no auth required, but with strict rate limiting)
	r.Group(func(r chi.Router) {
		r.Use(middleware.StrictRateLimit())
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.RefreshToken)

	})

	// Authenticated auth routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtManager, log))
		r.Post("/auth/logout", authHandler.Logout)
		r.Get("/auth/me", authHandler.Profile)
	})

	// Public API routes (no auth required for browsing)
	r.Group(func(r chi.Router) {
		// Public product browsing
		r.Get("/api/products", gatewayHandler.RouteRequest)
		r.Get("/api/products/{productId}", gatewayHandler.RouteRequest)

		// Public category browsing
		r.Get("/api/categories", gatewayHandler.RouteRequest)
		r.Get("/api/categories/{categoryId}", gatewayHandler.RouteRequest)
	})

	// Protected API routes (authentication required)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtManager, log))

		// Product management (admin only for create/update/delete)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin(log))
			r.Post("/api/products", gatewayHandler.RouteRequest)
			r.Put("/api/products/{productId}", gatewayHandler.RouteRequest)
			r.Delete("/api/products/{productId}", gatewayHandler.RouteRequest)

			// Category management (admin only)
			r.Post("/api/categories", gatewayHandler.RouteRequest)
			r.Put("/api/categories/{categoryId}", gatewayHandler.RouteRequest)
			r.Delete("/api/categories/{categoryId}", gatewayHandler.RouteRequest)
		})

		// Cart routes
		r.Get("/api/carts/{userId}", gatewayHandler.RouteRequest)
		r.Post("/api/carts/{userId}/items", gatewayHandler.RouteRequest)
		r.Put("/api/carts/{userId}/items/{productId}", gatewayHandler.RouteRequest)
		r.Delete("/api/carts/{userId}/items/{productId}", gatewayHandler.RouteRequest)
		r.Delete("/api/carts/{userId}", gatewayHandler.RouteRequest)

		// Order routes
		r.Get("/api/orders", gatewayHandler.RouteRequest)
		r.Post("/api/orders", gatewayHandler.RouteRequest)
		r.Get("/api/orders/{orderId}", gatewayHandler.RouteRequest)
		r.Put("/api/orders/{orderId}/status", gatewayHandler.RouteRequest)
		r.Post("/api/orders/{orderId}/cancel", gatewayHandler.RouteRequest)

		// Payment routes
		r.Get("/api/payments", gatewayHandler.RouteRequest)
		r.Post("/api/payments", gatewayHandler.RouteRequest)
		r.Get("/api/payments/{paymentId}", gatewayHandler.RouteRequest)
		r.Post("/api/payments/{paymentId}/capture", gatewayHandler.RouteRequest)
		r.Post("/api/payments/{paymentId}/cancel", gatewayHandler.RouteRequest)
		r.Post("/api/payments/{paymentId}/refund", gatewayHandler.RouteRequest)

		// Inventory routes (admin only)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin(log))
			r.Get("/api/inventory/stocks/{productId}", gatewayHandler.RouteRequest)
			r.Post("/api/inventory/stocks/{productId}/add", gatewayHandler.RouteRequest)
			r.Post("/api/inventory/stocks/{productId}/set", gatewayHandler.RouteRequest)
			r.Get("/api/inventory/stocks/low", gatewayHandler.RouteRequest)
			r.Get("/api/inventory/reservations/{orderId}", gatewayHandler.RouteRequest)
		})
	})

	return r
}
