package device

import "errors"

var (
	ErrDeviceNotFound    = errors.New("device not found")
	ErrInvalidMachineID  = errors.New("machine ID cannot be empty")
	ErrDeviceInactive    = errors.New("device is inactive")
	ErrDuplicateMachineID = errors.New("machine ID already registered")
)
