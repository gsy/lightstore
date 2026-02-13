package device

import (
	"time"

	"github.com/vending-machine/server/internal/domain/shared"
)

type DeviceStatus string

const (
	DeviceStatusActive   DeviceStatus = "active"
	DeviceStatusInactive DeviceStatus = "inactive"
)

// Device is the aggregate root for vending machine devices
type Device struct {
	id        shared.DeviceID
	machineID string // QR code identifier
	name      string
	location  string
	status    DeviceStatus
	createdAt time.Time
	updatedAt time.Time

	events []DomainEvent
}

// NewDevice creates a new device with validation
func NewDevice(machineID, name, location string) (*Device, error) {
	if machineID == "" {
		return nil, ErrInvalidMachineID
	}

	now := time.Now().UTC()
	d := &Device{
		id:        shared.NewDeviceID(),
		machineID: machineID,
		name:      name,
		location:  location,
		status:    DeviceStatusActive,
		createdAt: now,
		updatedAt: now,
	}

	d.events = append(d.events, DeviceRegistered{
		DeviceID:  d.id,
		MachineID: machineID,
	})

	return d, nil
}

// Reconstitute rebuilds a Device from persistence
func Reconstitute(
	id shared.DeviceID,
	machineID, name, location string,
	status DeviceStatus,
	createdAt, updatedAt time.Time,
) *Device {
	return &Device{
		id:        id,
		machineID: machineID,
		name:      name,
		location:  location,
		status:    status,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Getters
func (d *Device) ID() shared.DeviceID   { return d.id }
func (d *Device) MachineID() string     { return d.machineID }
func (d *Device) Name() string          { return d.name }
func (d *Device) Location() string      { return d.location }
func (d *Device) Status() DeviceStatus  { return d.status }
func (d *Device) CreatedAt() time.Time  { return d.createdAt }
func (d *Device) UpdatedAt() time.Time  { return d.updatedAt }

func (d *Device) IsActive() bool {
	return d.status == DeviceStatusActive
}

// Business methods

func (d *Device) Deactivate() {
	if d.status == DeviceStatusInactive {
		return
	}
	d.status = DeviceStatusInactive
	d.updatedAt = time.Now().UTC()
}

func (d *Device) Activate() {
	if d.status == DeviceStatusActive {
		return
	}
	d.status = DeviceStatusActive
	d.updatedAt = time.Now().UTC()
}

// PullEvents returns accumulated domain events and clears the slice
func (d *Device) PullEvents() []DomainEvent {
	events := d.events
	d.events = nil
	return events
}
