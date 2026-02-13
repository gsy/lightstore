package sku

import "github.com/vending-machine/server/internal/domain/shared"

// DomainEvent marker interface
type DomainEvent interface {
	domainEvent()
}

type SKUCreated struct {
	SKUID shared.SKUID
	Code  string
	Name  string
}

func (SKUCreated) domainEvent() {}

type SKUUpdated struct {
	SKUID shared.SKUID
	Name  string
}

func (SKUUpdated) domainEvent() {}

type SKUDeactivated struct {
	SKUID shared.SKUID
}

func (SKUDeactivated) domainEvent() {}
