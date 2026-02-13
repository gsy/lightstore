package adapters

import (
	"context"

	deviceapi "github.com/vending-machine/server/internal/device/api"
	"github.com/vending-machine/server/internal/transaction/app/ports"
)

// DeviceAdapter implements ports.DeviceReader using the device context API
type DeviceAdapter struct {
	reader deviceapi.DeviceReader
}

func NewDeviceAdapter(reader deviceapi.DeviceReader) *DeviceAdapter {
	if reader == nil {
		panic("nil DeviceReader")
	}
	return &DeviceAdapter{reader: reader}
}

func (a *DeviceAdapter) FindByMachineID(ctx context.Context, machineID string) (*ports.DeviceInfo, error) {
	view, err := a.reader.FindByMachineID(ctx, machineID)
	if err != nil {
		return nil, err
	}

	return &ports.DeviceInfo{
		ID:        view.ID,
		MachineID: view.MachineID,
		IsActive:  view.IsActive,
	}, nil
}
