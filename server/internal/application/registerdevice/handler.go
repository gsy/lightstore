package registerdevice

import (
	"context"
	"fmt"

	"github.com/vending-machine/server/internal/application/ports"
	"github.com/vending-machine/server/internal/domain/device"
)

// RegisterDeviceResult is the output DTO
type RegisterDeviceResult struct {
	DeviceID  string
	MachineID string
	IsNew     bool
}

// Handler orchestrates the device registration use case
type Handler struct {
	devices   device.DeviceRepository
	publisher ports.EventPublisher
}

func NewHandler(devices device.DeviceRepository, publisher ports.EventPublisher) *Handler {
	return &Handler{
		devices:   devices,
		publisher: publisher,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd RegisterDeviceCommand) (RegisterDeviceResult, error) {
	// Check if device already exists
	existing, _ := h.devices.FindByMachineID(ctx, cmd.MachineID)
	if existing != nil {
		return RegisterDeviceResult{
			DeviceID:  existing.ID().String(),
			MachineID: existing.MachineID(),
			IsNew:     false,
		}, nil
	}

	// Create new device
	dev, err := device.NewDevice(cmd.MachineID, cmd.Name, cmd.Location)
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
