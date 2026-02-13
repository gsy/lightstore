package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vending-machine/server/internal/application/registerdevice"
	"github.com/vending-machine/server/internal/application/submitdetection"
	"github.com/vending-machine/server/internal/domain/device"
	"github.com/vending-machine/server/internal/domain/sku"
)

type DeviceHandler struct {
	registerHandler  *registerdevice.Handler
	detectionHandler *submitdetection.Handler
	skuRepo          sku.SKURepository
}

func NewDeviceHandler(
	registerHandler *registerdevice.Handler,
	detectionHandler *submitdetection.Handler,
	skuRepo sku.SKURepository,
) *DeviceHandler {
	return &DeviceHandler{
		registerHandler:  registerHandler,
		detectionHandler: detectionHandler,
		skuRepo:          skuRepo,
	}
}

// Request/Response DTOs

type registerDeviceRequest struct {
	MachineID string `json:"machine_id" binding:"required"`
	Name      string `json:"name"`
	Location  string `json:"location"`
}

type detectedItemRequest struct {
	SKU        string    `json:"sku" binding:"required"`
	Confidence float64   `json:"confidence" binding:"required"`
	BBox       []float64 `json:"bbox"`
}

type submitDetectionRequest struct {
	DeviceID    string                `json:"device_id" binding:"required"`
	SessionID   string                `json:"session_id" binding:"required"`
	Items       []detectedItemRequest `json:"items" binding:"required"`
	TotalWeight float64               `json:"total_weight" binding:"required"`
}

type detectedItemResponse struct {
	SKU        string  `json:"sku"`
	Name       string  `json:"name"`
	PriceCents int64   `json:"price_cents"`
	Currency   string  `json:"currency"`
	Confidence float64 `json:"confidence"`
}

// Handlers

func (h *DeviceHandler) Register(c *gin.Context) {
	var req registerDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := registerdevice.RegisterDeviceCommand{
		MachineID: req.MachineID,
		Name:      req.Name,
		Location:  req.Location,
	}

	result, err := h.registerHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, device.ErrInvalidMachineID):
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

func (h *DeviceHandler) SubmitDetection(c *gin.Context) {
	var req submitDetectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []submitdetection.DetectedItemInput
	for _, item := range req.Items {
		items = append(items, submitdetection.DetectedItemInput{
			SKU:        item.SKU,
			Confidence: item.Confidence,
			BBox:       item.BBox,
		})
	}

	cmd := submitdetection.SubmitDetectionCommand{
		DeviceID:    req.DeviceID,
		SessionID:   req.SessionID,
		Items:       items,
		TotalWeight: req.TotalWeight,
	}

	result, err := h.detectionHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	var itemsResponse []detectedItemResponse
	for _, item := range result.Items {
		itemsResponse = append(itemsResponse, detectedItemResponse{
			SKU:        item.SKU,
			Name:       item.Name,
			PriceCents: item.PriceCents,
			Currency:   item.Currency,
			Confidence: item.Confidence,
		})
	}

	response := gin.H{
		"status":       "confirmed",
		"session_id":   result.SessionID,
		"items":        itemsResponse,
		"total_cents":  result.TotalCents,
		"currency":     result.Currency,
		"weight_match": result.WeightMatch,
	}

	if result.NeedsCloudML {
		response["status"] = "needs_verification"
		response["upload_image"] = true
	}

	c.JSON(http.StatusOK, response)
}

func (h *DeviceHandler) GetSKUs(c *gin.Context) {
	skus, err := h.skuRepo.FindAllActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var response []gin.H
	for _, s := range skus {
		response = append(response, gin.H{
			"code":             s.Code(),
			"name":             s.Name(),
			"weight_grams":     s.Weight().Grams(),
			"weight_tolerance": s.WeightTolerance(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"skus":  response,
		"count": len(response),
	})
}
