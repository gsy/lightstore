package valueobjects

import (
	"errors"

	"github.com/google/uuid"
)

// DeviceID is a strongly-typed ID for devices
type DeviceID struct {
	value uuid.UUID
}

func NewDeviceID() DeviceID {
	return DeviceID{value: uuid.New()}
}

func DeviceIDFrom(raw string) (DeviceID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return DeviceID{}, errors.New("invalid device ID format")
	}
	return DeviceID{value: id}, nil
}

func (d DeviceID) String() string { return d.value.String() }
func (d DeviceID) IsZero() bool   { return d.value == uuid.Nil }

// SKUID is a strongly-typed ID for SKUs
type SKUID struct {
	value uuid.UUID
}

func NewSKUID() SKUID {
	return SKUID{value: uuid.New()}
}

func SKUIDFrom(raw string) (SKUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return SKUID{}, errors.New("invalid SKU ID format")
	}
	return SKUID{value: id}, nil
}

func (s SKUID) String() string { return s.value.String() }
func (s SKUID) IsZero() bool   { return s.value == uuid.Nil }

// SessionID is a strongly-typed ID for sessions
type SessionID struct {
	value uuid.UUID
}

func NewSessionID() SessionID {
	return SessionID{value: uuid.New()}
}

func SessionIDFrom(raw string) (SessionID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return SessionID{}, errors.New("invalid session ID format")
	}
	return SessionID{value: id}, nil
}

func (s SessionID) String() string { return s.value.String() }
func (s SessionID) IsZero() bool   { return s.value == uuid.Nil }

// DetectionID is a strongly-typed ID for detections
type DetectionID struct {
	value uuid.UUID
}

func NewDetectionID() DetectionID {
	return DetectionID{value: uuid.New()}
}

func DetectionIDFrom(raw string) (DetectionID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return DetectionID{}, errors.New("invalid detection ID format")
	}
	return DetectionID{value: id}, nil
}

func (d DetectionID) String() string { return d.value.String() }
func (d DetectionID) IsZero() bool   { return d.value == uuid.Nil }

// TransactionID is a strongly-typed ID for transactions
type TransactionID struct {
	value uuid.UUID
}

func NewTransactionID() TransactionID {
	return TransactionID{value: uuid.New()}
}

func TransactionIDFrom(raw string) (TransactionID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return TransactionID{}, errors.New("invalid transaction ID format")
	}
	return TransactionID{value: id}, nil
}

func (t TransactionID) String() string { return t.value.String() }
func (t TransactionID) IsZero() bool   { return t.value == uuid.Nil }
