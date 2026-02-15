package valueobjects

import (
	"errors"
	"fmt"
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
