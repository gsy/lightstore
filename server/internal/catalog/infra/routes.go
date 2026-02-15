package infra

import "github.com/gin-gonic/gin"

// RegisterRoutes registers the catalog context routes
func (h *HTTPHandler) RegisterRoutes(rg *gin.RouterGroup) {
	skus := rg.Group("/skus")
	{
		skus.POST("", h.Create)
		skus.GET("", h.List)
		skus.GET("/active", h.ListActive)
		skus.GET("/:id", h.Get)
	}
}
