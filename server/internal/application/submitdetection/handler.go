package submitdetection

import (
	"context"
	"fmt"

	"github.com/vending-machine/server/internal/application/ports"
	"github.com/vending-machine/server/internal/domain/session"
	"github.com/vending-machine/server/internal/domain/shared"
	"github.com/vending-machine/server/internal/domain/sku"
)

const confidenceThreshold = 0.80

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
	SessionID      string
	Items          []DetectedItemOutput
	TotalCents     int64
	Currency       string
	WeightMatch    bool
	NeedsCloudML   bool
}

// Handler orchestrates the detection submission use case
type Handler struct {
	sessions  session.SessionRepository
	skus      sku.SKURepository
	publisher ports.EventPublisher
}

func NewHandler(
	sessions session.SessionRepository,
	skus sku.SKURepository,
	publisher ports.EventPublisher,
) *Handler {
	return &Handler{
		sessions:  sessions,
		skus:      skus,
		publisher: publisher,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd SubmitDetectionCommand) (SubmitDetectionResult, error) {
	// Parse session ID
	sessionID, err := shared.SessionIDFrom(cmd.SessionID)
	if err != nil {
		return SubmitDetectionResult{}, fmt.Errorf("invalid session ID: %w", err)
	}

	// Find session
	sess, err := h.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return SubmitDetectionResult{}, session.ErrSessionNotFound
	}

	if !sess.IsActive() {
		return SubmitDetectionResult{}, session.ErrSessionNotActive
	}

	// Enrich detected items with SKU details
	var detectedItems []session.DetectedItem
	var outputItems []DetectedItemOutput
	var expectedWeightGrams float64
	var needsCloudML bool
	var totalCents int64
	currency := "USD" // default

	for _, item := range cmd.Items {
		s, err := h.skus.FindByCode(ctx, item.SKU)
		if err != nil {
			needsCloudML = true
			continue
		}

		price := s.Price()
		detectedItem := session.NewDetectedItem(
			s.ID(),
			s.Code(),
			s.Name(),
			item.Confidence,
			price,
		)
		detectedItems = append(detectedItems, detectedItem)

		outputItems = append(outputItems, DetectedItemOutput{
			SKU:        s.Code(),
			Name:       s.Name(),
			PriceCents: price.Amount(),
			Currency:   price.Currency(),
			Confidence: item.Confidence,
		})

		expectedWeightGrams += s.Weight().Grams()
		totalCents += price.Amount()
		currency = price.Currency()

		if item.Confidence < confidenceThreshold {
			needsCloudML = true
		}
	}

	// Check weight tolerance
	measuredWeight, _ := shared.NewWeight(cmd.TotalWeight)
	expectedWeight, _ := shared.NewWeight(expectedWeightGrams)
	weightTolerance := 10.0 // grams
	weightMatch := expectedWeight.IsWithinTolerance(measuredWeight, weightTolerance)

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
