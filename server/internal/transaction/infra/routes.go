package infra

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all transaction context routes
func (h *HTTPHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Session routes
	sessions := r.Group("/session")
	{
		sessions.POST("/start", h.Start)
		sessions.GET("/:id", h.Get)
		sessions.POST("/:id/confirm", h.Confirm)
		sessions.POST("/:id/cancel", h.Cancel)
	}

	// Device detection route (used by ESP32 devices)
	device := r.Group("/device")
	{
		device.POST("/detection", h.SubmitDetection)
	}
}
