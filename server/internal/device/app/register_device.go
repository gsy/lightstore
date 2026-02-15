package app

import (
	"context"
	"fmt"

	"github.com/vending-machine/server/internal/device/domain"
	"github.com/vending-machine/server/internal/shared/events"
)

// EventPublisher is an output port for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event events.DomainEvent) error
}

// RegisterDeviceCommand is the input DTO for registering a device
type RegisterDeviceCommand struct {
	MachineID string
	Name      string
	Location  string
}

// RegisterDeviceResult is the output DTO
type RegisterDeviceResult struct {
	DeviceID  string
	MachineID string
	IsNew     bool
}

// RegisterDeviceHandler orchestrates the device registration use case
type RegisterDeviceHandler struct {
	devices   domain.DeviceRepository
	publisher EventPublisher
}

func NewRegisterDeviceHandler(devices domain.DeviceRepository, publisher EventPublisher) *RegisterDeviceHandler {
	if devices == nil {
		panic("nil DeviceRepository")
	}
	if publisher == nil {
		panic("nil EventPublisher")
	}
	return &RegisterDeviceHandler{
		devices:   devices,
		publisher: publisher,
	}
}

func (h *RegisterDeviceHandler) Handle(ctx context.Context, cmd RegisterDeviceCommand) (RegisterDeviceResult, error) {
	// Check if device already exists (idempotent)
	existing, _ := h.devices.FindByMachineID(ctx, cmd.MachineID)
	if existing != nil {
		return RegisterDeviceResult{
			DeviceID:  existing.ID().String(),
			MachineID: existing.MachineID(),
			IsNew:     false,
		}, nil
	}

	// Create new device
	dev, err := domain.NewDevice(cmd.MachineID, cmd.Name, cmd.Location)
	if err != nil {
		return RegisterDeviceResult{}, fmt.Errorf("invalid device: %w", err)
	}

	// Persist
	if err := h.devices.Save(ctx, dev); err != nil {
		return RegisterDeviceResult{}, fmt.Errorf("failed to save device: %w", err)
	}

	// Publish domain events
	for _, evt := range dev.PullEvents() {
		_ = h.publisher.Publish(ctx, evt)
	}

	return RegisterDeviceResult{
		DeviceID:  dev.ID().String(),
		MachineID: dev.MachineID(),
		IsNew:     true,
	}, nil
}
