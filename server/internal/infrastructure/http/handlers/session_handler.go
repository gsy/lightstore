package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vending-machine/server/internal/application/startsession"
	"github.com/vending-machine/server/internal/domain/device"
	"github.com/vending-machine/server/internal/domain/session"
	"github.com/vending-machine/server/internal/domain/shared"
)

type SessionHandler struct {
	startHandler *startsession.Handler
	sessionRepo  session.SessionRepository
}

func NewSessionHandler(
	startHandler *startsession.Handler,
	sessionRepo session.SessionRepository,
) *SessionHandler {
	return &SessionHandler{
		startHandler: startHandler,
		sessionRepo:  sessionRepo,
	}
}

// Request/Response DTOs

type startSessionRequest struct {
	MachineID string `json:"machine_id" binding:"required"`
	UserID    string `json:"user_id"`
}

type sessionItemResponse struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	PriceCents int64   `json:"price_cents"`
	Currency   string  `json:"currency"`
	Confidence float64 `json:"confidence"`
}

// Handlers

func (h *SessionHandler) Start(c *gin.Context) {
	var req startSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := startsession.StartSessionCommand{
		MachineID: req.MachineID,
		UserID:    req.UserID,
	}

	result, err := h.startHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, device.ErrDeviceNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		case errors.Is(err, device.ErrDeviceInactive):
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

func (h *SessionHandler) Get(c *gin.Context) {
	sessionID, err := shared.SessionIDFrom(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	sess, err := h.sessionRepo.FindByID(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var items []sessionItemResponse
	for _, item := range sess.DetectedItems() {
		items = append(items, sessionItemResponse{
			Code:       item.Code(),
			Name:       item.Name(),
			PriceCents: item.Price().Amount(),
			Currency:   item.Price().Currency(),
			Confidence: item.Confidence(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"session": gin.H{
			"id":         sess.ID().String(),
			"device_id":  sess.DeviceID().String(),
			"status":     string(sess.Status()),
			"created_at": sess.CreatedAt(),
			"expires_at": sess.ExpiresAt(),
		},
		"items":       items,
		"total_cents": sess.TotalAmount().Amount(),
		"currency":    sess.TotalAmount().Currency(),
	})
}

func (h *SessionHandler) Confirm(c *gin.Context) {
	sessionID, err := shared.SessionIDFrom(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	var req struct {
		PaymentRef string `json:"payment_ref"`
	}
	_ = c.ShouldBindJSON(&req)

	sess, err := h.sessionRepo.FindByID(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := sess.Confirm(req.PaymentRef); err != nil {
		switch {
		case errors.Is(err, session.ErrSessionNotActive):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "session not active"})
		case errors.Is(err, session.ErrNoItemsDetected):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "no items detected"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	if err := h.sessionRepo.Save(c.Request.Context(), sess); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "completed",
		"message": "purchase confirmed",
	})
}

func (h *SessionHandler) Cancel(c *gin.Context) {
	sessionID, err := shared.SessionIDFrom(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	sess, err := h.sessionRepo.FindByID(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := sess.Cancel(req.Reason); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if err := h.sessionRepo.Save(c.Request.Context(), sess); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "cancelled",
		"message": "session cancelled",
	})
}
