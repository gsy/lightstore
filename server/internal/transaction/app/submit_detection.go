package app

import (
	"context"
	"fmt"

	"github.com/vending-machine/server/internal/shared/policy"
	"github.com/vending-machine/server/internal/shared/valueobjects"
	"github.com/vending-machine/server/internal/transaction/app/ports"
	"github.com/vending-machine/server/internal/transaction/domain"
)

// DetectedItemInput represents a detected item from the device
type DetectedItemInput struct {
	SKU        string
	Confidence float64
	BBox       []float64
}

// SubmitDetectionCommand is the input DTO for submitting detection results
type SubmitDetectionCommand struct {
	DeviceID    string
	SessionID   string
	Items       []DetectedItemInput
	TotalWeight float64
}

// DetectedItemOutput represents an enriched detected item
type DetectedItemOutput struct {
	SKU        string
	Name       string
	PriceCents int64
	Currency   string
	Confidence float64
}

// SubmitDetectionResult is the output DTO
type SubmitDetectionResult struct {
	SessionID    string
	Items        []DetectedItemOutput
	TotalCents   int64
	Currency     string
	WeightMatch  bool
	NeedsCloudML bool
}

// SubmitDetectionHandler orchestrates the detection submission use case
type SubmitDetectionHandler struct {
	sessions  domain.SessionRepository
	catalog   ports.CatalogReader
	publisher eventPublisher
	policy    policy.DetectionPolicy
}

func NewSubmitDetectionHandler(
	sessions domain.SessionRepository,
	catalog ports.CatalogReader,
	publisher eventPublisher,
) *SubmitDetectionHandler {
	if sessions == nil {
		panic("nil SessionRepository")
	}
	if catalog == nil {
		panic("nil CatalogReader")
	}
	if publisher == nil {
		panic("nil EventPublisher")
	}
	return &SubmitDetectionHandler{
		sessions:  sessions,
		catalog:   catalog,
		publisher: publisher,
		policy:    policy.DefaultDetectionPolicy(),
	}
}

// NewSubmitDetectionHandlerWithPolicy creates a handler with a custom detection policy
func NewSubmitDetectionHandlerWithPolicy(
	sessions domain.SessionRepository,
	catalog ports.CatalogReader,
	publisher eventPublisher,
	detectionPolicy policy.DetectionPolicy,
) *SubmitDetectionHandler {
	if sessions == nil {
		panic("nil SessionRepository")
	}
	if catalog == nil {
		panic("nil CatalogReader")
	}
	if publisher == nil {
		panic("nil EventPublisher")
	}
	return &SubmitDetectionHandler{
		sessions:  sessions,
		catalog:   catalog,
		publisher: publisher,
		policy:    detectionPolicy,
	}
}

func (h *SubmitDetectionHandler) Handle(ctx context.Context, cmd SubmitDetectionCommand) (SubmitDetectionResult, error) {
	// Parse session ID
	sessionID, err := valueobjects.SessionIDFrom(cmd.SessionID)
	if err != nil {
		return SubmitDetectionResult{}, fmt.Errorf("invalid session ID: %w", err)
	}

	// Find session
	sess, err := h.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return SubmitDetectionResult{}, domain.ErrSessionNotFound
	}

	if !sess.IsActive() {
		return SubmitDetectionResult{}, domain.ErrSessionNotActive
	}

	// Enrich detected items with SKU details from catalog context
	var detectedItems []domain.DetectedItem
	var outputItems []DetectedItemOutput
	var expectedWeightGrams float64
	var needsCloudML bool
	var totalCents int64
	currency := "USD" // default

	for _, item := range cmd.Items {
		skuInfo, err := h.catalog.FindSKUByCode(ctx, item.SKU)
		if err != nil {
			needsCloudML = true
			continue
		}

		skuID, _ := valueobjects.SKUIDFrom(skuInfo.ID)
		price, _ := valueobjects.NewMoney(skuInfo.PriceCents, skuInfo.Currency)

		detectedItem := domain.NewDetectedItem(
			skuID,
			skuInfo.Code,
			skuInfo.Name,
			item.Confidence,
			price,
		)
		detectedItems = append(detectedItems, detectedItem)

		outputItems = append(outputItems, DetectedItemOutput{
			SKU:        skuInfo.Code,
			Name:       skuInfo.Name,
			PriceCents: skuInfo.PriceCents,
			Currency:   skuInfo.Currency,
			Confidence: item.Confidence,
		})

		expectedWeightGrams += skuInfo.WeightGrams
		totalCents += skuInfo.PriceCents
		currency = skuInfo.Currency

		if !h.policy.IsConfidenceAcceptable(item.Confidence) {
			needsCloudML = true
		}
	}

	// Check weight tolerance using policy
	measuredWeight, _ := valueobjects.NewWeight(cmd.TotalWeight)
	expectedWeight, _ := valueobjects.NewWeight(expectedWeightGrams)
	weightMatch := h.policy.IsWeightMatch(expectedWeight, measuredWeight)

	if !weightMatch {
		needsCloudML = true
	}

	// Record detection in session
	if err := sess.RecordDetection(detectedItems, measuredWeight); err != nil {
		return SubmitDetectionResult{}, fmt.Errorf("failed to record detection: %w", err)
	}

	// Persist
	if err := h.sessions.Save(ctx, sess); err != nil {
		return SubmitDetectionResult{}, fmt.Errorf("failed to save session: %w", err)
	}

	// Publish domain events
	for _, evt := range sess.PullEvents() {
		_ = h.publisher.Publish(ctx, evt)
	}

	return SubmitDetectionResult{
		SessionID:    sess.ID().String(),
		Items:        outputItems,
		TotalCents:   totalCents,
		Currency:     currency,
		WeightMatch:  weightMatch,
		NeedsCloudML: needsCloudML,
	}, nil
}
