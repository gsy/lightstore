package app

import (
	"context"

	"github.com/vending-machine/server/internal/device/domain"
	"github.com/vending-machine/server/internal/shared/valueobjects"
)

// DeviceQueryService provides read-only access to devices
type DeviceQueryService struct {
	repo domain.DeviceRepository
}

func NewDeviceQueryService(repo domain.DeviceRepository) *DeviceQueryService {
	return &DeviceQueryService{repo: repo}
}

func (s *DeviceQueryService) FindByID(ctx context.Context, id string) (*domain.Device, error) {
	deviceID, err := valueobjects.DeviceIDFrom(id)
	if err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, deviceID)
}

func (s *DeviceQueryService) FindByMachineID(ctx context.Context, machineID string) (*domain.Device, error) {
	return s.repo.FindByMachineID(ctx, machineID)
}
