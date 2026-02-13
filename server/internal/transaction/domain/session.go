package domain

import (
	"time"

	"github.com/vending-machine/server/internal/shared/events"
	"github.com/vending-machine/server/internal/shared/valueobjects"
)

type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusCancelled SessionStatus = "cancelled"
	SessionStatusExpired   SessionStatus = "expired"
)

// Session is the aggregate root for a customer interaction session
type Session struct {
	id            valueobjects.SessionID
	deviceID      valueobjects.DeviceID
	userID        string
	status        SessionStatus
	detectedItems []DetectedItem
	totalWeight   valueobjects.Weight
	totalAmount   valueobjects.Money
	createdAt     time.Time
	expiresAt     time.Time
	completedAt   *time.Time

	domainEvents []events.DomainEvent
}

// NewSession creates a new session when user scans QR code
func NewSession(deviceID valueobjects.DeviceID, userID string, expirationMinutes int) (*Session, error) {
	if deviceID.IsZero() {
		return nil, ErrInvalidDeviceID
	}

	now := time.Now().UTC()
	s := &Session{
		id:            valueobjects.NewSessionID(),
		deviceID:      deviceID,
		userID:        userID,
		status:        SessionStatusActive,
		detectedItems: []DetectedItem{},
		createdAt:     now,
		expiresAt:     now.Add(time.Duration(expirationMinutes) * time.Minute),
	}

	s.domainEvents = append(s.domainEvents, NewSessionStarted(s.id, deviceID, userID))

	return s, nil
}

// Reconstitute rebuilds a Session from persistence
func Reconstitute(
	id valueobjects.SessionID,
	deviceID valueobjects.DeviceID,
	userID string,
	status SessionStatus,
	detectedItems []DetectedItem,
	totalWeight valueobjects.Weight,
	totalAmount valueobjects.Money,
	createdAt, expiresAt time.Time,
	completedAt *time.Time,
) *Session {
	return &Session{
		id:            id,
		deviceID:      deviceID,
		userID:        userID,
		status:        status,
		detectedItems: detectedItems,
		totalWeight:   totalWeight,
		totalAmount:   totalAmount,
		createdAt:     createdAt,
		expiresAt:     expiresAt,
		completedAt:   completedAt,
	}
}

// Getters
func (s *Session) ID() valueobjects.SessionID       { return s.id }
func (s *Session) DeviceID() valueobjects.DeviceID  { return s.deviceID }
func (s *Session) UserID() string                   { return s.userID }
func (s *Session) Status() SessionStatus            { return s.status }
func (s *Session) DetectedItems() []DetectedItem    { return append([]DetectedItem{}, s.detectedItems...) }
func (s *Session) TotalWeight() valueobjects.Weight { return s.totalWeight }
func (s *Session) TotalAmount() valueobjects.Money  { return s.totalAmount }
func (s *Session) CreatedAt() time.Time             { return s.createdAt }
func (s *Session) ExpiresAt() time.Time             { return s.expiresAt }
func (s *Session) CompletedAt() *time.Time          { return s.completedAt }

func (s *Session) IsActive() bool {
	return s.status == SessionStatusActive && time.Now().Before(s.expiresAt)
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.expiresAt)
}

// Business methods

// RecordDetection records items detected by the device
func (s *Session) RecordDetection(items []DetectedItem, totalWeight valueobjects.Weight) error {
	if !s.IsActive() {
		return ErrSessionNotActive
	}
	if s.IsExpired() {
		s.status = SessionStatusExpired
		return ErrSessionExpired
	}

	s.detectedItems = items
	s.totalWeight = totalWeight

	// Calculate total amount
	var total valueobjects.Money
	for i, item := range items {
		if i == 0 {
			total = item.Price()
		} else {
			var err error
			total, err = total.Add(item.Price())
			if err != nil {
				return err
			}
		}
	}
	s.totalAmount = total

	s.domainEvents = append(s.domainEvents, NewItemsDetected(s.id, len(items), totalWeight.Grams()))

	return nil
}

// Confirm completes the session after payment
func (s *Session) Confirm(paymentRef string) error {
	if !s.IsActive() {
		return ErrSessionNotActive
	}
	if len(s.detectedItems) == 0 {
		return ErrNoItemsDetected
	}

	now := time.Now().UTC()
	s.status = SessionStatusCompleted
	s.completedAt = &now

	s.domainEvents = append(s.domainEvents, NewSessionCompleted(s.id, paymentRef))

	return nil
}

// Cancel cancels the session
func (s *Session) Cancel(reason string) error {
	if s.status == SessionStatusCompleted {
		return ErrSessionAlreadyCompleted
	}

	now := time.Now().UTC()
	s.status = SessionStatusCancelled
	s.completedAt = &now

	s.domainEvents = append(s.domainEvents, NewSessionCancelled(s.id, reason))

	return nil
}

// PullEvents returns accumulated domain events and clears the slice
func (s *Session) PullEvents() []events.DomainEvent {
	evts := s.domainEvents
	s.domainEvents = nil
	return evts
}
