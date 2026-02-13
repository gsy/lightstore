package domain

import (
	"context"

	"github.com/vending-machine/server/internal/shared/valueobjects"
)

// SessionRepository is the PORT interface defined by the domain
type SessionRepository interface {
	Save(ctx context.Context, session *Session) error
	FindByID(ctx context.Context, id valueobjects.SessionID) (*Session, error)
	FindActiveByDeviceID(ctx context.Context, deviceID valueobjects.DeviceID) (*Session, error)
}
