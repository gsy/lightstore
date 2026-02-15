package app

import (
	"context"
	"fmt"

	"github.com/vending-machine/server/internal/catalog/domain"
	"github.com/vending-machine/server/internal/shared/events"
)

// EventPublisher is an output port for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event events.DomainEvent) error
}

// CreateSKUCommand is the input DTO for creating a SKU
type CreateSKUCommand struct {
	Code            string
	Name            string
	PriceCents      int64
	Currency        string
	WeightGrams     float64
	WeightTolerance float64
	ImageURL        string
}

// CreateSKUResult is the output DTO
type CreateSKUResult struct {
	SKUID string
}

// CreateSKUHandler orchestrates the SKU creation use case
type CreateSKUHandler struct {
	skus      domain.SKURepository
	publisher EventPublisher
}

func NewCreateSKUHandler(skus domain.SKURepository, publisher EventPublisher) *CreateSKUHandler {
	if skus == nil {
		panic("nil SKURepository")
	}
	if publisher == nil {
		panic("nil EventPublisher")
	}
	return &CreateSKUHandler{
		skus:      skus,
		publisher: publisher,
	}
}

func (h *CreateSKUHandler) Handle(ctx context.Context, cmd CreateSKUCommand) (CreateSKUResult, error) {
	// Check for duplicate code
	existing, _ := h.skus.FindByCode(ctx, cmd.Code)
	if existing != nil {
		return CreateSKUResult{}, domain.ErrDuplicateSKUCode
	}

	// Create the aggregate
	s, err := domain.NewSKU(cmd.Code, cmd.Name, cmd.PriceCents, cmd.Currency, cmd.WeightGrams)
	if err != nil {
		return CreateSKUResult{}, fmt.Errorf("invalid SKU: %w", err)
	}

	// Update optional fields
	if cmd.WeightTolerance > 0 || cmd.ImageURL != "" {
		err = s.Update(cmd.Name, cmd.PriceCents, cmd.Currency, cmd.WeightGrams, cmd.WeightTolerance, cmd.ImageURL)
		if err != nil {
			return CreateSKUResult{}, fmt.Errorf("failed to update SKU: %w", err)
		}
	}

	// Persist
	if err := h.skus.Save(ctx, s); err != nil {
		return CreateSKUResult{}, fmt.Errorf("failed to save SKU: %w", err)
	}

	// Publish domain events
	for _, evt := range s.PullEvents() {
		_ = h.publisher.Publish(ctx, evt)
	}

	return CreateSKUResult{SKUID: s.ID().String()}, nil
}
