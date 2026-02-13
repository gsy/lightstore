package session

import (
	"time"

	"github.com/vending-machine/server/internal/domain/shared"
)

type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusCancelled SessionStatus = "cancelled"
	SessionStatusExpired   SessionStatus = "expired"
)

// DetectedItem is a value object representing a detected SKU
type DetectedItem struct {
	skuID      shared.SKUID
	code       string
	name       string
	confidence float64
	price      shared.Money
}

func NewDetectedItem(skuID shared.SKUID, code, name string, confidence float64, price shared.Money) DetectedItem {
	return DetectedItem{
		skuID:      skuID,
		code:       code,
		name:       name,
		confidence: confidence,
		price:      price,
	}
}

func (d DetectedItem) SKUID() shared.SKUID  { return d.skuID }
func (d DetectedItem) Code() string         { return d.code }
func (d DetectedItem) Name() string         { return d.name }
func (d DetectedItem) Confidence() float64  { return d.confidence }
func (d DetectedItem) Price() shared.Money  { return d.price }

// Session is the aggregate root for a customer interaction session
type Session struct {
	id           shared.SessionID
	deviceID     shared.DeviceID
	userID       string
	status       SessionStatus
	detectedItems []DetectedItem
	totalWeight  shared.Weight
	totalAmount  shared.Money
	createdAt    time.Time
	expiresAt    time.Time
	completedAt  *time.Time

	events []DomainEvent
}

// NewSession creates a new session when user scans QR code
func NewSession(deviceID shared.DeviceID, userID string, expirationMinutes int) (*Session, error) {
	if deviceID.IsZero() {
		return nil, ErrInvalidDeviceID
	}

	now := time.Now().UTC()
	s := &Session{
		id:           shared.NewSessionID(),
		deviceID:     deviceID,
		userID:       userID,
		status:       SessionStatusActive,
		detectedItems: []DetectedItem{},
		createdAt:    now,
		expiresAt:    now.Add(time.Duration(expirationMinutes) * time.Minute),
	}

	s.events = append(s.events, SessionStarted{
		SessionID: s.id,
		DeviceID:  deviceID,
		UserID:    userID,
	})

	return s, nil
}

// Reconstitute rebuilds a Session from persistence
func Reconstitute(
	id shared.SessionID,
	deviceID shared.DeviceID,
	userID string,
	status SessionStatus,
	detectedItems []DetectedItem,
	totalWeight shared.Weight,
	totalAmount shared.Money,
	createdAt, expiresAt time.Time,
	completedAt *time.Time,
) *Session {
	return &Session{
		id:           id,
		deviceID:     deviceID,
		userID:       userID,
		status:       status,
		detectedItems: detectedItems,
		totalWeight:  totalWeight,
		totalAmount:  totalAmount,
		createdAt:    createdAt,
		expiresAt:    expiresAt,
		completedAt:  completedAt,
	}
}

// Getters
func (s *Session) ID() shared.SessionID        { return s.id }
func (s *Session) DeviceID() shared.DeviceID   { return s.deviceID }
func (s *Session) UserID() string              { return s.userID }
func (s *Session) Status() SessionStatus       { return s.status }
func (s *Session) DetectedItems() []DetectedItem { return append([]DetectedItem{}, s.detectedItems...) }
func (s *Session) TotalWeight() shared.Weight  { return s.totalWeight }
func (s *Session) TotalAmount() shared.Money   { return s.totalAmount }
func (s *Session) CreatedAt() time.Time        { return s.createdAt }
func (s *Session) ExpiresAt() time.Time        { return s.expiresAt }
func (s *Session) CompletedAt() *time.Time     { return s.completedAt }

func (s *Session) IsActive() bool {
	return s.status == SessionStatusActive && time.Now().Before(s.expiresAt)
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.expiresAt)
}

// Business methods

// RecordDetection records items detected by the device
func (s *Session) RecordDetection(items []DetectedItem, totalWeight shared.Weight) error {
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
	var total shared.Money
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

	s.events = append(s.events, ItemsDetected{
		SessionID:   s.id,
		ItemCount:   len(items),
		TotalWeight: totalWeight.Grams(),
	})

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

	s.events = append(s.events, SessionCompleted{
		SessionID:  s.id,
		PaymentRef: paymentRef,
	})

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

	s.events = append(s.events, SessionCancelled{
		SessionID: s.id,
		Reason:    reason,
	})

	return nil
}

// PullEvents returns accumulated domain events and clears the slice
func (s *Session) PullEvents() []DomainEvent {
	events := s.events
	s.events = nil
	return events
}
