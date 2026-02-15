package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vending-machine/server/internal/shared/events"
	"github.com/vending-machine/server/internal/shared/valueobjects"
	"github.com/vending-machine/server/internal/transaction/app/ports"
	"github.com/vending-machine/server/internal/transaction/domain"
)

const defaultSessionExpirationMinutes = 30

// StartSessionCommand is the input DTO for starting a session
type StartSessionCommand struct {
	MachineID string
	UserID    string
}

// StartSessionResult is the output DTO
type StartSessionResult struct {
	SessionID string
	DeviceID  string
	ExpiresAt time.Time
}

// Errors for start session use case
var (
	ErrDeviceNotFound = errors.New("device not found")
	ErrDeviceInactive = errors.New("device is inactive")
)

// eventPublisher is a local interface for publishing domain events
type eventPublisher interface {
	Publish(ctx context.Context, event events.DomainEvent) error
}

// StartSessionHandler orchestrates the session start use case
type StartSessionHandler struct {
	devices   ports.DeviceReader
	sessions  domain.SessionRepository
	publisher eventPublisher
}

func NewStartSessionHandler(
	devices ports.DeviceReader,
	sessions domain.SessionRepository,
	publisher eventPublisher,
) *StartSessionHandler {
	if devices == nil {
		panic("nil DeviceReader")
	}
	if sessions == nil {
		panic("nil SessionRepository")
	}
	if publisher == nil {
		panic("nil EventPublisher")
	}
	return &StartSessionHandler{
		devices:   devices,
		sessions:  sessions,
		publisher: publisher,
	}
}

func (h *StartSessionHandler) Handle(ctx context.Context, cmd StartSessionCommand) (StartSessionResult, error) {
	// Find device by machine ID using the cross-context port
	dev, err := h.devices.FindByMachineID(ctx, cmd.MachineID)
	if err != nil {
		return StartSessionResult{}, ErrDeviceNotFound
	}

	if !dev.IsActive {
		return StartSessionResult{}, ErrDeviceInactive
	}

	// Parse device ID
	deviceID, err := valueobjects.DeviceIDFrom(dev.ID)
	if err != nil {
		return StartSessionResult{}, fmt.Errorf("invalid device ID: %w", err)
	}

	// Create new session
	sess, err := domain.NewSession(deviceID, cmd.UserID, defaultSessionExpirationMinutes)
	if err != nil {
		return StartSessionResult{}, fmt.Errorf("failed to create session: %w", err)
	}

	// Persist
	if err := h.sessions.Save(ctx, sess); err != nil {
		return StartSessionResult{}, fmt.Errorf("failed to save session: %w", err)
	}

	// Publish domain events
	for _, evt := range sess.PullEvents() {
		_ = h.publisher.Publish(ctx, evt)
	}

	return StartSessionResult{
		SessionID: sess.ID().String(),
		DeviceID:  dev.ID,
		ExpiresAt: sess.ExpiresAt(),
	}, nil
}
