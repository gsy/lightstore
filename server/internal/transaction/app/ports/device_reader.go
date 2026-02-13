package ports

import "context"

// DeviceInfo is a DTO representing device information needed by transaction context
type DeviceInfo struct {
	ID        string
	MachineID string
	IsActive  bool
}

// DeviceReader is an input port for reading device context data.
// This port is defined by the transaction context (consumer) and
// implemented by an adapter that calls the device context API.
type DeviceReader interface {
	FindByMachineID(ctx context.Context, machineID string) (*DeviceInfo, error)
}
