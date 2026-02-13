package errors

import "errors"

// Shared domain errors
var (
	ErrInvalidConfidenceThreshold = errors.New("confidence threshold must be between 0 and 1")
	ErrInvalidWeightTolerance     = errors.New("weight tolerance cannot be negative")
)
