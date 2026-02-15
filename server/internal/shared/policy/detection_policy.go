package policy

import (
	"github.com/vending-machine/server/internal/shared/errors"
	"github.com/vending-machine/server/internal/shared/valueobjects"
)

// DetectionPolicy encapsulates the business rules for detection validation.
// This is a Value Object that defines thresholds and tolerances for the
// ML detection process.
type DetectionPolicy struct {
	confidenceThreshold  float64 // Minimum confidence to accept detection (0.0-1.0)
	weightToleranceGrams float64 // Maximum weight difference in grams
}

// DefaultDetectionPolicy returns the standard detection policy
func DefaultDetectionPolicy() DetectionPolicy {
	return DetectionPolicy{
		confidenceThreshold:  0.80,
		weightToleranceGrams: 10.0,
	}
}

// NewDetectionPolicy creates a custom detection policy with validation
func NewDetectionPolicy(confidenceThreshold, weightToleranceGrams float64) (DetectionPolicy, error) {
	if confidenceThreshold < 0 || confidenceThreshold > 1 {
		return DetectionPolicy{}, errors.ErrInvalidConfidenceThreshold
	}
	if weightToleranceGrams < 0 {
		return DetectionPolicy{}, errors.ErrInvalidWeightTolerance
	}
	return DetectionPolicy{
		confidenceThreshold:  confidenceThreshold,
		weightToleranceGrams: weightToleranceGrams,
	}, nil
}

// ConfidenceThreshold returns the minimum confidence level
func (p DetectionPolicy) ConfidenceThreshold() float64 {
	return p.confidenceThreshold
}

// WeightToleranceGrams returns the weight tolerance in grams
func (p DetectionPolicy) WeightToleranceGrams() float64 {
	return p.weightToleranceGrams
}

// IsConfidenceAcceptable checks if a confidence value meets the threshold
func (p DetectionPolicy) IsConfidenceAcceptable(confidence float64) bool {
	return confidence >= p.confidenceThreshold
}

// IsWeightMatch checks if expected and measured weights are within tolerance
func (p DetectionPolicy) IsWeightMatch(expected, measured valueobjects.Weight) bool {
	return expected.IsWithinTolerance(measured, p.weightToleranceGrams)
}
