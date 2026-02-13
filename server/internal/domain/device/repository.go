package device

import (
	"context"

	"github.com/vending-machine/server/internal/domain/shared"
)

// DeviceRepository is the PORT interface defined by the domain
type DeviceRepository interface {
	Save(ctx context.Context, device *Device) error
	FindByID(ctx context.Context, id shared.DeviceID) (*Device, error)
	FindByMachineID(ctx context.Context, machineID string) (*Device, error)
}
