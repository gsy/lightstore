package messaging

import (
	"context"

	"github.com/vending-machine/server/internal/application/ports"
	"github.com/vending-machine/server/internal/pkg/logger"
)

// NoOpEventPublisher is a placeholder event publisher that just logs events
type NoOpEventPublisher struct{}

func NewNoOpEventPublisher() *NoOpEventPublisher {
	return &NoOpEventPublisher{}
}

func (p *NoOpEventPublisher) Publish(ctx context.Context, event ports.Event) error {
	logger.Debug("Domain event published (no-op)", "event", event)
	return nil
}
