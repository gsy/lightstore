package shared

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// Money is a Value Object representing monetary amounts
type Money struct {
	amount   int64  // stored in cents
	currency string // ISO 4217 code
}

func NewMoney(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, errors.New("money amount cannot be negative")
	}
	if len(currency) != 3 {
		return Money{}, errors.New("currency must be a 3-letter ISO code")
	}
	return Money{amount: amount, currency: currency}, nil
}

func (m Money) Amount() int64    { return m.amount }
func (m Money) Currency() string { return m.currency }

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot add %s to %s", other.currency, m.currency)
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", float64(m.amount)/100, m.currency)
}

// Weight is a Value Object representing weight in grams
type Weight struct {
	grams float64
}

func NewWeight(grams float64) (Weight, error) {
	if grams < 0 {
		return Weight{}, errors.New("weight cannot be negative")
	}
	return Weight{grams: grams}, nil
}

func (w Weight) Grams() float64 { return w.grams }

func (w Weight) IsWithinTolerance(other Weight, tolerance float64) bool {
	diff := w.grams - other.grams
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

func (w Weight) Add(other Weight) Weight {
	return Weight{grams: w.grams + other.grams}
}

// DeviceID is a strongly-typed ID
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

// SKUID is a strongly-typed ID
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

// SessionID is a strongly-typed ID
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

// DetectionID is a strongly-typed ID
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

// TransactionID is a strongly-typed ID
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
