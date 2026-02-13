package session

import "github.com/vending-machine/server/internal/domain/shared"

type DomainEvent interface {
	domainEvent()
}

type SessionStarted struct {
	SessionID shared.SessionID
	DeviceID  shared.DeviceID
	UserID    string
}

func (SessionStarted) domainEvent() {}

type ItemsDetected struct {
	SessionID   shared.SessionID
	ItemCount   int
	TotalWeight float64
}

func (ItemsDetected) domainEvent() {}

type SessionCompleted struct {
	SessionID  shared.SessionID
	PaymentRef string
}

func (SessionCompleted) domainEvent() {}

type SessionCancelled struct {
	SessionID shared.SessionID
	Reason    string
}

func (SessionCancelled) domainEvent() {}
