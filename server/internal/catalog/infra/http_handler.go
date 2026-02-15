package infra

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vending-machine/server/internal/catalog/app"
	"github.com/vending-machine/server/internal/catalog/domain"
)

type HTTPHandler struct {
	createHandler *app.CreateSKUHandler
	queryService  *app.SKUQueryService
}

func NewHTTPHandler(
	createHandler *app.CreateSKUHandler,
	queryService *app.SKUQueryService,
) *HTTPHandler {
	return &HTTPHandler{
		createHandler: createHandler,
		queryService:  queryService,
	}
}

// Request/Response DTOs (HTTP layer only)

type createSKURequest struct {
	Code            string  `json:"code" binding:"required"`
	Name            string  `json:"name" binding:"required"`
	PriceCents      int64   `json:"price_cents" binding:"required"`
	Currency        string  `json:"currency"`
	WeightGrams     float64 `json:"weight_grams" binding:"required"`
	WeightTolerance float64 `json:"weight_tolerance"`
	ImageURL        string  `json:"image_url"`
}

type skuResponse struct {
	ID              string  `json:"id"`
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	PriceCents      int64   `json:"price_cents"`
	Currency        string  `json:"currency"`
	WeightGrams     float64 `json:"weight_grams"`
	WeightTolerance float64 `json:"weight_tolerance"`
	ImageURL        string  `json:"image_url,omitempty"`
	Active          bool    `json:"active"`
}

// Handlers

func (h *HTTPHandler) Create(c *gin.Context) {
	var req createSKURequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	cmd := app.CreateSKUCommand{
		Code:            req.Code,
		Name:            req.Name,
		PriceCents:      req.PriceCents,
		Currency:        currency,
		WeightGrams:     req.WeightGrams,
		WeightTolerance: req.WeightTolerance,
		ImageURL:        req.ImageURL,
	}

	result, err := h.createHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDuplicateSKUCode):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrInvalidSKUName),
			errors.Is(err, domain.ErrInvalidSKUPrice),
			errors.Is(err, domain.ErrInvalidSKUWeight):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      result.SKUID,
		"message": "SKU created",
	})
}

func (h *HTTPHandler) Get(c *gin.Context) {
	s, err := h.queryService.FindByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, domain.ErrSKUNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "SKU not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	c.JSON(http.StatusOK, toSKUResponse(s))
}

func (h *HTTPHandler) List(c *gin.Context) {
	skus, err := h.queryService.FindAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var response []skuResponse
	for _, s := range skus {
		response = append(response, toSKUResponse(s))
	}

	c.JSON(http.StatusOK, gin.H{
		"skus":  response,
		"count": len(response),
	})
}

func (h *HTTPHandler) ListActive(c *gin.Context) {
	skus, err := h.queryService.FindAllActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	var response []skuResponse
	for _, s := range skus {
		response = append(response, toSKUResponse(s))
	}

	c.JSON(http.StatusOK, gin.H{
		"skus":  response,
		"count": len(response),
	})
}

func toSKUResponse(s *domain.SKU) skuResponse {
	return skuResponse{
		ID:              s.ID().String(),
		Code:            s.Code(),
		Name:            s.Name(),
		PriceCents:      s.Price().Amount(),
		Currency:        s.Price().Currency(),
		WeightGrams:     s.Weight().Grams(),
		WeightTolerance: s.WeightTolerance(),
		ImageURL:        s.ImageURL(),
		Active:          s.IsActive(),
	}
}
