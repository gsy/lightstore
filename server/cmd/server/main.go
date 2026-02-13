package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vending-machine/server/internal/application/createsku"
	"github.com/vending-machine/server/internal/application/registerdevice"
	"github.com/vending-machine/server/internal/application/startsession"
	"github.com/vending-machine/server/internal/application/submitdetection"
	httpserver "github.com/vending-machine/server/internal/infrastructure/http"
	"github.com/vending-machine/server/internal/infrastructure/http/handlers"
	"github.com/vending-machine/server/internal/infrastructure/messaging"
	"github.com/vending-machine/server/internal/infrastructure/persistence/postgres"
	"github.com/vending-machine/server/internal/pkg/logger"
)

func main() {
	// Setup logging
	logger.Init(logger.WithLevel(slog.LevelDebug))

	logger.Info("Starting Vending Machine Server (DDD Architecture)")

	// Load config
	port := getEnv("PORT", "8080")
	databaseURL := getEnv("DATABASE_URL", "postgres://vending:vending@localhost:5432/vending?sslmode=disable")

	// Connect to database
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		logger.Fatal("Failed to ping database", "error", err)
	}
	logger.Info("Connected to database")

	// Run migrations
	if err := postgres.RunMigrations(pool); err != nil {
		logger.Fatal("Failed to run migrations", "error", err)
	}

	// =========================================================================
	// Infrastructure Layer - Adapters
	// =========================================================================

	// Repository adapters (implement domain interfaces)
	skuRepo := postgres.NewPostgresSKURepository(pool)
	deviceRepo := postgres.NewPostgresDeviceRepository(pool)
	sessionRepo := postgres.NewPostgresSessionRepository(pool)

	// Output port adapters
	eventPublisher := messaging.NewNoOpEventPublisher()

	// =========================================================================
	// Application Layer - Use Cases
	// =========================================================================

	createSKUHandler := createsku.NewHandler(skuRepo, eventPublisher)
	registerDeviceHandler := registerdevice.NewHandler(deviceRepo, eventPublisher)
	startSessionHandler := startsession.NewHandler(deviceRepo, sessionRepo, eventPublisher)
	submitDetectionHandler := submitdetection.NewHandler(sessionRepo, skuRepo, eventPublisher)

	// =========================================================================
	// Interface Layer - HTTP Handlers
	// =========================================================================

	skuHandler := handlers.NewSKUHandler(createSKUHandler, skuRepo)
	deviceHandler := handlers.NewDeviceHandler(registerDeviceHandler, submitDetectionHandler, skuRepo)
	sessionHandler := handlers.NewSessionHandler(startSessionHandler, sessionRepo)

	// Create router
	router := httpserver.NewRouter(skuHandler, deviceHandler, sessionHandler)

	// Create server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
