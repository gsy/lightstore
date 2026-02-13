package registerdevice

// RegisterDeviceCommand is the input DTO for registering a device
type RegisterDeviceCommand struct {
	MachineID string
	Name      string
	Location  string
}
