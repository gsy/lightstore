package http

import (
	"github.com/gin-gonic/gin"

	"github.com/vending-machine/server/internal/infrastructure/http/handlers"
)

func NewRouter(
	skuHandler *handlers.SKUHandler,
	deviceHandler *handlers.DeviceHandler,
	sessionHandler *handlers.SessionHandler,
) *gin.Engine {
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Device endpoints (ESP32 -> Server)
		device := v1.Group("/device")
		{
			device.POST("/register", deviceHandler.Register)
			device.POST("/detection", deviceHandler.SubmitDetection)
			device.GET("/skus", deviceHandler.GetSKUs)
		}

		// Session endpoints (App -> Server)
		session := v1.Group("/session")
		{
			session.POST("/start", sessionHandler.Start)
			session.GET("/:id", sessionHandler.Get)
			session.POST("/:id/confirm", sessionHandler.Confirm)
			session.POST("/:id/cancel", sessionHandler.Cancel)
		}

		// SKU management endpoints (Admin)
		skus := v1.Group("/skus")
		{
			skus.GET("", skuHandler.List)
			skus.GET("/active", skuHandler.ListActive)
			skus.GET("/:id", skuHandler.Get)
			skus.POST("", skuHandler.Create)
		}
	}

	return router
}
