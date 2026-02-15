package messaging

import (
	"context"

	"github.com/vending-machine/server/internal/pkg/logger"
	"github.com/vending-machine/server/internal/shared/events"
)

// NoOpEventPublisher is a placeholder event publisher that just logs events.
// In production, this would be replaced with a real message broker adapter
// (e.g., Kafka, RabbitMQ, AWS SNS/SQS).
type NoOpEventPublisher struct{}

func NewNoOpEventPublisher() *NoOpEventPublisher {
	return &NoOpEventPublisher{}
}

// Publish logs the domain event (no-op implementation)
func (p *NoOpEventPublisher) Publish(ctx context.Context, event events.DomainEvent) error {
	logger.Debug("Domain event published (no-op)",
		"event_name", event.EventName(),
		"occurred_at", event.OccurredAt(),
	)
	return nil
}
