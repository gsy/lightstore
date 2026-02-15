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

	// Catalog context
	catalogapi "github.com/vending-machine/server/internal/catalog/api"
	catalogapp "github.com/vending-machine/server/internal/catalog/app"
	cataloginfra "github.com/vending-machine/server/internal/catalog/infra"

	// Device context
	deviceapi "github.com/vending-machine/server/internal/device/api"
	deviceapp "github.com/vending-machine/server/internal/device/app"
	deviceinfra "github.com/vending-machine/server/internal/device/infra"

	// Transaction context
	transactionapp "github.com/vending-machine/server/internal/transaction/app"
	transactioninfra "github.com/vending-machine/server/internal/transaction/infra"
	transactionadapters "github.com/vending-machine/server/internal/transaction/infra/adapters"

	// Platform
	platformhttp "github.com/vending-machine/server/internal/platform/http"
	"github.com/vending-machine/server/internal/platform/messaging"
	"github.com/vending-machine/server/internal/platform/postgres"

	// Shared
	"github.com/vending-machine/server/internal/pkg/logger"
)

func main() {
	// Setup logging
	logger.Init(logger.WithLevel(slog.LevelDebug))

	logger.Info("Starting Vending Machine Server (Modular DDD Architecture)")

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
	// Shared Infrastructure
	// =========================================================================

	eventPublisher := messaging.NewNoOpEventPublisher()

	// =========================================================================
	// Catalog Bounded Context
	// =========================================================================

	// Infrastructure layer
	skuRepo := cataloginfra.NewPostgresSKURepository(pool)

	// API layer (cross-context communication)
	skuReader := catalogapi.NewSKUReaderAdapter(skuRepo)

	// Application layer
	createSKUHandler := catalogapp.NewCreateSKUHandler(skuRepo, eventPublisher)
	skuQueryService := catalogapp.NewSKUQueryService(skuRepo)

	// HTTP handler
	catalogHandler := cataloginfra.NewHTTPHandler(createSKUHandler, skuQueryService)

	// =========================================================================
	// Device Bounded Context
	// =========================================================================

	// Infrastructure layer
	deviceRepo := deviceinfra.NewPostgresDeviceRepository(pool)

	// API layer (cross-context communication)
	deviceReader := deviceapi.NewDeviceReaderAdapter(deviceRepo)

	// Application layer
	registerDeviceHandler := deviceapp.NewRegisterDeviceHandler(deviceRepo, eventPublisher)

	// HTTP handler (with cross-context SKU reader)
	deviceHandler := deviceinfra.NewHTTPHandler(registerDeviceHandler, skuReader)

	// =========================================================================
	// Transaction Bounded Context
	// =========================================================================

	// Infrastructure layer
	sessionRepo := transactioninfra.NewPostgresSessionRepository(pool)

	// Cross-context adapters (implements transaction's ports using other contexts' APIs)
	deviceAdapter := transactionadapters.NewDeviceAdapter(deviceReader)
	catalogAdapter := transactionadapters.NewCatalogAdapter(skuReader)

	// Application layer
	startSessionHandler := transactionapp.NewStartSessionHandler(deviceAdapter, sessionRepo, eventPublisher)
	submitDetectionHandler := transactionapp.NewSubmitDetectionHandler(sessionRepo, catalogAdapter, eventPublisher)
	confirmSessionHandler := transactionapp.NewConfirmSessionHandler(sessionRepo, eventPublisher)
	cancelSessionHandler := transactionapp.NewCancelSessionHandler(sessionRepo, eventPublisher)
	sessionQueryService := transactionapp.NewSessionQueryService(sessionRepo)

	// HTTP handler
	transactionHandler := transactioninfra.NewHTTPHandler(
		startSessionHandler,
		submitDetectionHandler,
		confirmSessionHandler,
		cancelSessionHandler,
		sessionQueryService,
	)

	// =========================================================================
	// HTTP Router (composes all context routes)
	// =========================================================================

	router := platformhttp.NewRouter(catalogHandler, deviceHandler, transactionHandler)

	// Create server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router.Engine(),
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
