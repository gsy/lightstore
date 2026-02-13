package domain

import (
	"github.com/vending-machine/server/internal/shared/events"
	"github.com/vending-machine/server/internal/shared/valueobjects"
)

type SessionStarted struct {
	events.BaseEvent
	SessionID valueobjects.SessionID
	DeviceID  valueobjects.DeviceID
	UserID    string
}

func NewSessionStarted(sessionID valueobjects.SessionID, deviceID valueobjects.DeviceID, userID string) SessionStarted {
	return SessionStarted{
		BaseEvent: events.NewBaseEvent(),
		SessionID: sessionID,
		DeviceID:  deviceID,
		UserID:    userID,
	}
}

func (SessionStarted) EventName() string { return "SessionStarted" }

type ItemsDetected struct {
	events.BaseEvent
	SessionID   valueobjects.SessionID
	ItemCount   int
	TotalWeight float64
}

func NewItemsDetected(sessionID valueobjects.SessionID, itemCount int, totalWeight float64) ItemsDetected {
	return ItemsDetected{
		BaseEvent:   events.NewBaseEvent(),
		SessionID:   sessionID,
		ItemCount:   itemCount,
		TotalWeight: totalWeight,
	}
}

func (ItemsDetected) EventName() string { return "ItemsDetected" }

type SessionCompleted struct {
	events.BaseEvent
	SessionID  valueobjects.SessionID
	PaymentRef string
}

func NewSessionCompleted(sessionID valueobjects.SessionID, paymentRef string) SessionCompleted {
	return SessionCompleted{
		BaseEvent:  events.NewBaseEvent(),
		SessionID:  sessionID,
		PaymentRef: paymentRef,
	}
}

func (SessionCompleted) EventName() string { return "SessionCompleted" }

type SessionCancelled struct {
	events.BaseEvent
	SessionID valueobjects.SessionID
	Reason    string
}

func NewSessionCancelled(sessionID valueobjects.SessionID, reason string) SessionCancelled {
	return SessionCancelled{
		BaseEvent: events.NewBaseEvent(),
		SessionID: sessionID,
		Reason:    reason,
	}
}

func (SessionCancelled) EventName() string { return "SessionCancelled" }
