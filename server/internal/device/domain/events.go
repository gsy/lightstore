package domain

import (
	"github.com/vending-machine/server/internal/shared/events"
	"github.com/vending-machine/server/internal/shared/valueobjects"
)

type DeviceRegistered struct {
	events.BaseEvent
	DeviceID  valueobjects.DeviceID
	MachineID string
}

func NewDeviceRegistered(deviceID valueobjects.DeviceID, machineID string) DeviceRegistered {
	return DeviceRegistered{
		BaseEvent: events.NewBaseEvent(),
		DeviceID:  deviceID,
		MachineID: machineID,
	}
}

func (DeviceRegistered) EventName() string { return "DeviceRegistered" }
