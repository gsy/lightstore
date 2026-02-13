package ports

import "context"

// Event is a generic marker interface for all domain events
type Event interface{}

// EventPublisher is an OUTPUT PORT for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}
