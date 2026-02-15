package valueobjects

import "errors"

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
