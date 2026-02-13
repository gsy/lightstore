package createsku

import (
	"context"
	"fmt"

	"github.com/vending-machine/server/internal/application/ports"
	"github.com/vending-machine/server/internal/domain/sku"
)

// CreateSKUResult is the output DTO
type CreateSKUResult struct {
	SKUID string
}

// Handler orchestrates the SKU creation use case
type Handler struct {
	skus      sku.SKURepository
	publisher ports.EventPublisher
}

func NewHandler(skus sku.SKURepository, publisher ports.EventPublisher) *Handler {
	return &Handler{
		skus:      skus,
		publisher: publisher,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd CreateSKUCommand) (CreateSKUResult, error) {
	// Check for duplicate code
	existing, _ := h.skus.FindByCode(ctx, cmd.Code)
	if existing != nil {
		return CreateSKUResult{}, sku.ErrDuplicateSKUCode
	}

	// Create the aggregate
	s, err := sku.NewSKU(cmd.Code, cmd.Name, cmd.PriceCents, cmd.Currency, cmd.WeightGrams)
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
