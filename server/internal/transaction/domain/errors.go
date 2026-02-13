package domain

import "errors"

var (
	ErrSessionNotFound         = errors.New("session not found")
	ErrInvalidDeviceID         = errors.New("invalid device ID")
	ErrSessionNotActive        = errors.New("session is not active")
	ErrSessionExpired          = errors.New("session has expired")
	ErrSessionAlreadyCompleted = errors.New("session already completed")
	ErrNoItemsDetected         = errors.New("no items detected in session")
)
