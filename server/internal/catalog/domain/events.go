package domain

import (
	"github.com/vending-machine/server/internal/shared/events"
	"github.com/vending-machine/server/internal/shared/valueobjects"
)

type SKUCreated struct {
	events.BaseEvent
	SKUID valueobjects.SKUID
	Code  string
	Name  string
}

func NewSKUCreated(id valueobjects.SKUID, code, name string) SKUCreated {
	return SKUCreated{
		BaseEvent: events.NewBaseEvent(),
		SKUID:     id,
		Code:      code,
		Name:      name,
	}
}

func (SKUCreated) EventName() string { return "SKUCreated" }

type SKUUpdated struct {
	events.BaseEvent
	SKUID valueobjects.SKUID
	Name  string
}

func NewSKUUpdated(id valueobjects.SKUID, name string) SKUUpdated {
	return SKUUpdated{
		BaseEvent: events.NewBaseEvent(),
		SKUID:     id,
		Name:      name,
	}
}

func (SKUUpdated) EventName() string { return "SKUUpdated" }

type SKUDeactivated struct {
	events.BaseEvent
	SKUID valueobjects.SKUID
}

func NewSKUDeactivated(id valueobjects.SKUID) SKUDeactivated {
	return SKUDeactivated{
		BaseEvent: events.NewBaseEvent(),
		SKUID:     id,
	}
}

func (SKUDeactivated) EventName() string { return "SKUDeactivated" }
