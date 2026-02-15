package infra

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vending-machine/server/internal/catalog/api"
	"github.com/vending-machine/server/internal/device/app"
	"github.com/vending-machine/server/internal/device/domain"
)

type HTTPHandler struct {
	registerHandler *app.RegisterDeviceHandler
	skuReader       api.SKUReader // Cross-context read
}

func NewHTTPHandler(
	registerHandler *app.RegisterDeviceHandler,
	skuReader api.SKUReader,
) *HTTPHandler {
	return &HTTPHandler{
		registerHandler: registerHandler,
		skuReader:       skuReader,
	}
}

// Request/Response DTOs

type registerDeviceRequest struct {
	MachineID string `json:"machine_id" binding:"required"`
	Name      string `json:"name"`
	Location  string `json:"location"`
}

// Handlers

func (h *HTTPHandler) Register(c *gin.Context) {
	var req registerDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := app.RegisterDeviceCommand{
		MachineID: req.MachineID,
		Name:      req.Name,
		Location:  req.Location,
	}

	result, err := h.registerHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidMachineID):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	status := http.StatusCreated
	message := "device registered"
	if !result.IsNew {
		status = http.StatusOK
		message = "device already registered"
	}

	c.JSON(status, gin.H{
		"id":         result.DeviceID,
		"machine_id": result.MachineID,
		"message":    message,
	})
}

// GetSKUs returns active SKUs for device ML model sync
// This is a cross-context read using the Catalog API
func (h *HTTPHandler) GetSKUs(c *gin.Context) {
	skus, err := h.skuReader.FindAllActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var response []gin.H
	for _, s := range skus {
		response = append(response, gin.H{
			"code":             s.Code,
			"name":             s.Name,
			"weight_grams":     s.WeightGrams,
			"weight_tolerance": s.WeightTolerance,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"skus":  response,
		"count": len(response),
	})
}
