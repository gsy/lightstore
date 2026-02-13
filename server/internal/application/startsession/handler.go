package startsession

import (
	"context"
	"fmt"
	"time"

	"github.com/vending-machine/server/internal/application/ports"
	"github.com/vending-machine/server/internal/domain/device"
	"github.com/vending-machine/server/internal/domain/session"
)

const defaultSessionExpirationMinutes = 30

// StartSessionResult is the output DTO
type StartSessionResult struct {
	SessionID string
	DeviceID  string
	ExpiresAt time.Time
}

// Handler orchestrates the session start use case
type Handler struct {
	devices   device.DeviceRepository
	sessions  session.SessionRepository
	publisher ports.EventPublisher
}

func NewHandler(
	devices device.DeviceRepository,
	sessions session.SessionRepository,
	publisher ports.EventPublisher,
) *Handler {
	return &Handler{
		devices:   devices,
		sessions:  sessions,
		publisher: publisher,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd StartSessionCommand) (StartSessionResult, error) {
	// Find device by machine ID
	dev, err := h.devices.FindByMachineID(ctx, cmd.MachineID)
	if err != nil {
		return StartSessionResult{}, device.ErrDeviceNotFound
	}

	if !dev.IsActive() {
		return StartSessionResult{}, device.ErrDeviceInactive
	}

	// Create new session
	sess, err := session.NewSession(dev.ID(), cmd.UserID, defaultSessionExpirationMinutes)
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
		DeviceID:  dev.ID().String(),
		ExpiresAt: sess.ExpiresAt(),
	}, nil
}
