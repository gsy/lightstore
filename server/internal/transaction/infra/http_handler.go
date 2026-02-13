package infra

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vending-machine/server/internal/transaction/app"
	"github.com/vending-machine/server/internal/transaction/domain"
)

// HTTPHandler handles HTTP requests for the transaction context
type HTTPHandler struct {
	startHandler     *app.StartSessionHandler
	submitHandler    *app.SubmitDetectionHandler
	confirmHandler   *app.ConfirmSessionHandler
	cancelHandler    *app.CancelSessionHandler
	queryService     *app.SessionQueryService
}

func NewHTTPHandler(
	startHandler *app.StartSessionHandler,
	submitHandler *app.SubmitDetectionHandler,
	confirmHandler *app.ConfirmSessionHandler,
	cancelHandler *app.CancelSessionHandler,
	queryService *app.SessionQueryService,
) *HTTPHandler {
	return &HTTPHandler{
		startHandler:   startHandler,
		submitHandler:  submitHandler,
		confirmHandler: confirmHandler,
		cancelHandler:  cancelHandler,
		queryService:   queryService,
	}
}

// Request/Response DTOs

type startSessionRequest struct {
	MachineID string `json:"machine_id" binding:"required"`
	UserID    string `json:"user_id"`
}

type submitDetectionRequest struct {
	DeviceID    string                `json:"device_id" binding:"required"`
	SessionID   string                `json:"session_id" binding:"required"`
	Items       []detectedItemRequest `json:"items" binding:"required"`
	TotalWeight float64               `json:"total_weight"`
}

type detectedItemRequest struct {
	SKU        string    `json:"sku" binding:"required"`
	Confidence float64   `json:"confidence"`
	BBox       []float64 `json:"bbox"`
}

type sessionItemResponse struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	PriceCents int64   `json:"price_cents"`
	Currency   string  `json:"currency"`
	Confidence float64 `json:"confidence"`
}

// Handlers

func (h *HTTPHandler) Start(c *gin.Context) {
	var req startSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := app.StartSessionCommand{
		MachineID: req.MachineID,
		UserID:    req.UserID,
	}

	result, err := h.startHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrDeviceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		case errors.Is(err, app.ErrDeviceInactive):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "device is inactive"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id": result.SessionID,
		"device_id":  result.DeviceID,
		"expires_at": result.ExpiresAt,
		"message":    "session started, place items on scale",
	})
}

func (h *HTTPHandler) SubmitDetection(c *gin.Context) {
	var req submitDetectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []app.DetectedItemInput
	for _, item := range req.Items {
		items = append(items, app.DetectedItemInput{
			SKU:        item.SKU,
			Confidence: item.Confidence,
			BBox:       item.BBox,
		})
	}

	cmd := app.SubmitDetectionCommand{
		DeviceID:    req.DeviceID,
		SessionID:   req.SessionID,
		Items:       items,
		TotalWeight: req.TotalWeight,
	}

	result, err := h.submitHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		case errors.Is(err, domain.ErrSessionNotActive):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "session not active"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	var outputItems []sessionItemResponse
	for _, item := range result.Items {
		outputItems = append(outputItems, sessionItemResponse{
			Code:       item.SKU,
			Name:       item.Name,
			PriceCents: item.PriceCents,
			Currency:   item.Currency,
			Confidence: item.Confidence,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":    result.SessionID,
		"items":         outputItems,
		"total_cents":   result.TotalCents,
		"currency":      result.Currency,
		"weight_match":  result.WeightMatch,
		"needs_cloud_ml": result.NeedsCloudML,
	})
}

func (h *HTTPHandler) Get(c *gin.Context) {
	sessionID := c.Param("id")

	view, err := h.queryService.FindByID(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var items []sessionItemResponse
	for _, item := range view.Items {
		items = append(items, sessionItemResponse{
			Code:       item.Code,
			Name:       item.Name,
			PriceCents: item.PriceCents,
			Currency:   item.Currency,
			Confidence: item.Confidence,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"session": gin.H{
			"id":         view.ID,
			"device_id":  view.DeviceID,
			"status":     view.Status,
			"created_at": view.CreatedAt,
			"expires_at": view.ExpiresAt,
		},
		"items":       items,
		"total_cents": view.TotalCents,
		"currency":    view.Currency,
	})
}

func (h *HTTPHandler) Confirm(c *gin.Context) {
	var req struct {
		PaymentRef string `json:"payment_ref"`
	}
	_ = c.ShouldBindJSON(&req)

	cmd := app.ConfirmSessionCommand{
		SessionID:  c.Param("id"),
		PaymentRef: req.PaymentRef,
	}

	result, err := h.confirmHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		case errors.Is(err, domain.ErrSessionNotActive):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "session not active"})
		case errors.Is(err, domain.ErrNoItemsDetected):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "no items detected"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "completed",
		"message":     "purchase confirmed",
		"session_id":  result.SessionID,
		"total_cents": result.TotalCents,
		"currency":    result.Currency,
	})
}

func (h *HTTPHandler) Cancel(c *gin.Context) {
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	cmd := app.CancelSessionCommand{
		SessionID: c.Param("id"),
		Reason:    req.Reason,
	}

	result, err := h.cancelHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		case errors.Is(err, domain.ErrSessionAlreadyCompleted):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "session already completed"})
		default:
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "cancelled",
		"message":    "session cancelled",
		"session_id": result.SessionID,
	})
}
