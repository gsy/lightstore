package app

import (
	"context"
	"fmt"

	"github.com/vending-machine/server/internal/shared/valueobjects"
	"github.com/vending-machine/server/internal/transaction/domain"
)

// CancelSessionCommand is the input DTO for cancelling a session
type CancelSessionCommand struct {
	SessionID string
	Reason    string
}

// CancelSessionResult is the output DTO
type CancelSessionResult struct {
	SessionID string
	Reason    string
}

// CancelSessionHandler orchestrates the session cancellation use case
type CancelSessionHandler struct {
	sessions  domain.SessionRepository
	publisher eventPublisher
}

func NewCancelSessionHandler(sessions domain.SessionRepository, publisher eventPublisher) *CancelSessionHandler {
	if sessions == nil {
		panic("nil SessionRepository")
	}
	if publisher == nil {
		panic("nil EventPublisher")
	}
	return &CancelSessionHandler{
		sessions:  sessions,
		publisher: publisher,
	}
}

func (h *CancelSessionHandler) Handle(ctx context.Context, cmd CancelSessionCommand) (CancelSessionResult, error) {
	sessionID, err := valueobjects.SessionIDFrom(cmd.SessionID)
	if err != nil {
		return CancelSessionResult{}, fmt.Errorf("invalid session ID: %w", err)
	}

	sess, err := h.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return CancelSessionResult{}, domain.ErrSessionNotFound
	}

	if err := sess.Cancel(cmd.Reason); err != nil {
		return CancelSessionResult{}, err
	}

	if err := h.sessions.Save(ctx, sess); err != nil {
		return CancelSessionResult{}, fmt.Errorf("failed to save session: %w", err)
	}

	// Publish domain events
	for _, evt := range sess.PullEvents() {
		_ = h.publisher.Publish(ctx, evt)
	}

	return CancelSessionResult{
		SessionID: sess.ID().String(),
		Reason:    cmd.Reason,
	}, nil
}
