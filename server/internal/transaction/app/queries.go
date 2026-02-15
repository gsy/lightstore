package app

import (
	"context"

	"github.com/vending-machine/server/internal/shared/valueobjects"
	"github.com/vending-machine/server/internal/transaction/domain"
)

// SessionView is a read-only view of a session
type SessionView struct {
	ID          string
	DeviceID    string
	UserID      string
	Status      string
	Items       []SessionItemView
	TotalCents  int64
	Currency    string
	TotalWeight float64
	CreatedAt   string
	ExpiresAt   string
	CompletedAt *string
}

// SessionItemView is a read-only view of a detected item
type SessionItemView struct {
	SKUID      string
	Code       string
	Name       string
	Confidence float64
	PriceCents int64
	Currency   string
}

// SessionQueryService provides read-only access to sessions
type SessionQueryService struct {
	sessions domain.SessionRepository
}

func NewSessionQueryService(sessions domain.SessionRepository) *SessionQueryService {
	if sessions == nil {
		panic("nil SessionRepository")
	}
	return &SessionQueryService{sessions: sessions}
}

func (s *SessionQueryService) FindByID(ctx context.Context, id string) (*SessionView, error) {
	sessionID, err := valueobjects.SessionIDFrom(id)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	sess, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return s.toView(sess), nil
}

func (s *SessionQueryService) FindActiveByDeviceID(ctx context.Context, deviceID string) (*SessionView, error) {
	devID, err := valueobjects.DeviceIDFrom(deviceID)
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}

	sess, err := s.sessions.FindActiveByDeviceID(ctx, devID)
	if err != nil {
		return nil, err
	}

	return s.toView(sess), nil
}

func (s *SessionQueryService) toView(sess *domain.Session) *SessionView {
	var items []SessionItemView
	for _, item := range sess.DetectedItems() {
		items = append(items, SessionItemView{
			SKUID:      item.SKUID().String(),
			Code:       item.Code(),
			Name:       item.Name(),
			Confidence: item.Confidence(),
			PriceCents: item.Price().Amount(),
			Currency:   item.Price().Currency(),
		})
	}

	var completedAt *string
	if sess.CompletedAt() != nil {
		t := sess.CompletedAt().Format("2006-01-02T15:04:05Z07:00")
		completedAt = &t
	}

	return &SessionView{
		ID:          sess.ID().String(),
		DeviceID:    sess.DeviceID().String(),
		UserID:      sess.UserID(),
		Status:      string(sess.Status()),
		Items:       items,
		TotalCents:  sess.TotalAmount().Amount(),
		Currency:    sess.TotalAmount().Currency(),
		TotalWeight: sess.TotalWeight().Grams(),
		CreatedAt:   sess.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		ExpiresAt:   sess.ExpiresAt().Format("2006-01-02T15:04:05Z07:00"),
		CompletedAt: completedAt,
	}
}
