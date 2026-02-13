package session

import (
	"context"

	"github.com/vending-machine/server/internal/domain/shared"
)

// SessionRepository is the PORT interface defined by the domain
type SessionRepository interface {
	Save(ctx context.Context, session *Session) error
	FindByID(ctx context.Context, id shared.SessionID) (*Session, error)
	FindActiveByDeviceID(ctx context.Context, deviceID shared.DeviceID) (*Session, error)
}
