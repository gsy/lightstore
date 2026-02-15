package api

import (
	"context"

	"github.com/vending-machine/server/internal/device/domain"
	"github.com/vending-machine/server/internal/shared/valueobjects"
)

// DeviceView is a read-only DTO exposed to other bounded contexts
type DeviceView struct {
	ID        string
	MachineID string
	Name      string
	Location  string
	IsActive  bool
}

// DeviceReader is the interface other contexts use to read device data.
// This prevents direct domain coupling between bounded contexts.
type DeviceReader interface {
	FindByMachineID(ctx context.Context, machineID string) (*DeviceView, error)
	FindByID(ctx context.Context, id string) (*DeviceView, error)
}

// DeviceReaderAdapter implements DeviceReader using the domain repository
type DeviceReaderAdapter struct {
	repo domain.DeviceRepository
}

func NewDeviceReaderAdapter(repo domain.DeviceRepository) *DeviceReaderAdapter {
	return &DeviceReaderAdapter{repo: repo}
}

func (a *DeviceReaderAdapter) FindByMachineID(ctx context.Context, machineID string) (*DeviceView, error) {
	device, err := a.repo.FindByMachineID(ctx, machineID)
	if err != nil {
		return nil, err
	}
	return toDeviceView(device), nil
}

func (a *DeviceReaderAdapter) FindByID(ctx context.Context, id string) (*DeviceView, error) {
	deviceID, err := valueobjects.DeviceIDFrom(id)
	if err != nil {
		return nil, err
	}
	device, err := a.repo.FindByID(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	return toDeviceView(device), nil
}

func toDeviceView(d *domain.Device) *DeviceView {
	return &DeviceView{
		ID:        d.ID().String(),
		MachineID: d.MachineID(),
		Name:      d.Name(),
		Location:  d.Location(),
		IsActive:  d.IsActive(),
	}
}
