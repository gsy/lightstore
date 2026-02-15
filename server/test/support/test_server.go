package support

import (
	"context"
	"net/http/httptest"

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
)

// StartTestServer creates and starts a test HTTP server with all dependencies wired
func StartTestServer(pool *pgxpool.Pool) *httptest.Server {
	// Shared infrastructure
	eventPublisher := messaging.NewNoOpEventPublisher()

	// =========================================================================
	// Catalog Bounded Context
	// =========================================================================
	skuRepo := cataloginfra.NewPostgresSKURepository(pool)
	skuReader := catalogapi.NewSKUReaderAdapter(skuRepo)
	createSKUHandler := catalogapp.NewCreateSKUHandler(skuRepo, eventPublisher)
	skuQueryService := catalogapp.NewSKUQueryService(skuRepo)
	catalogHandler := cataloginfra.NewHTTPHandler(createSKUHandler, skuQueryService)

	// =========================================================================
	// Device Bounded Context
	// =========================================================================
	deviceRepo := deviceinfra.NewPostgresDeviceRepository(pool)
	deviceReader := deviceapi.NewDeviceReaderAdapter(deviceRepo)
	registerDeviceHandler := deviceapp.NewRegisterDeviceHandler(deviceRepo, eventPublisher)
	deviceHandler := deviceinfra.NewHTTPHandler(registerDeviceHandler, skuReader)

	// =========================================================================
	// Transaction Bounded Context
	// =========================================================================
	sessionRepo := transactioninfra.NewPostgresSessionRepository(pool)
	deviceAdapter := transactionadapters.NewDeviceAdapter(deviceReader)
	catalogAdapter := transactionadapters.NewCatalogAdapter(skuReader)
	startSessionHandler := transactionapp.NewStartSessionHandler(deviceAdapter, sessionRepo, eventPublisher)
	submitDetectionHandler := transactionapp.NewSubmitDetectionHandler(sessionRepo, catalogAdapter, eventPublisher)
	confirmSessionHandler := transactionapp.NewConfirmSessionHandler(sessionRepo, eventPublisher)
	cancelSessionHandler := transactionapp.NewCancelSessionHandler(sessionRepo, eventPublisher)
	sessionQueryService := transactionapp.NewSessionQueryService(sessionRepo)
	transactionHandler := transactioninfra.NewHTTPHandler(
		startSessionHandler,
		submitDetectionHandler,
		confirmSessionHandler,
		cancelSessionHandler,
		sessionQueryService,
	)

	// =========================================================================
	// HTTP Router
	// =========================================================================
	router := platformhttp.NewRouter(catalogHandler, deviceHandler, transactionHandler)

	return httptest.NewServer(router.Engine())
}

// ConnectTestDB connects to the test database
func ConnectTestDB(databaseURL string) (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), databaseURL)
}
