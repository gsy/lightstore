package infra

import "github.com/gin-gonic/gin"

// RegisterRoutes registers the device context routes
func (h *HTTPHandler) RegisterRoutes(rg *gin.RouterGroup) {
	device := rg.Group("/device")
	{
		device.POST("/register", h.Register)
		device.GET("/skus", h.GetSKUs)
	}
}
