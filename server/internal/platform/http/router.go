package http

import (
	"github.com/gin-gonic/gin"

	cataloginfra "github.com/vending-machine/server/internal/catalog/infra"
	deviceinfra "github.com/vending-machine/server/internal/device/infra"
	transactioninfra "github.com/vending-machine/server/internal/transaction/infra"
)

// Router composes all bounded context routes into a single Gin engine
type Router struct {
	catalogHandler     *cataloginfra.HTTPHandler
	deviceHandler      *deviceinfra.HTTPHandler
	transactionHandler *transactioninfra.HTTPHandler
}

// NewRouter creates a new router that composes all context handlers
func NewRouter(
	catalogHandler *cataloginfra.HTTPHandler,
	deviceHandler *deviceinfra.HTTPHandler,
	transactionHandler *transactioninfra.HTTPHandler,
) *Router {
	return &Router{
		catalogHandler:     catalogHandler,
		deviceHandler:      deviceHandler,
		transactionHandler: transactionHandler,
	}
}

// Engine returns a configured Gin engine with all routes registered
func (r *Router) Engine() *gin.Engine {
	engine := gin.Default()

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := engine.Group("/api/v1")
	{
		// Register all context routes
		r.catalogHandler.RegisterRoutes(v1)
		r.deviceHandler.RegisterRoutes(v1)
		r.transactionHandler.RegisterRoutes(v1)
	}

	return engine
}
