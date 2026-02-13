package device

import "github.com/vending-machine/server/internal/domain/shared"

type DomainEvent interface {
	domainEvent()
}

type DeviceRegistered struct {
	DeviceID  shared.DeviceID
	MachineID string
}

func (DeviceRegistered) domainEvent() {}
