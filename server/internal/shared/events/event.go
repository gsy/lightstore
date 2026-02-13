package events

import "time"

// DomainEvent is the base interface for all domain events.
// Domain events represent facts that have occurred in the domain.
// They are immutable and named in past tense.
type DomainEvent interface {
	// EventName returns the event's name (e.g., "SKUCreated", "SessionStarted")
	EventName() string
	// OccurredAt returns when the event occurred
	OccurredAt() time.Time
}

// BaseEvent provides common fields for all domain events.
// Embed this in concrete event types.
type BaseEvent struct {
	occurredAt time.Time
}

// NewBaseEvent creates a BaseEvent with the current timestamp
func NewBaseEvent() BaseEvent {
	return BaseEvent{occurredAt: time.Now().UTC()}
}

// OccurredAt returns when the event occurred
func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}
