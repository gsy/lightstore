package app

import (
	"context"
	"fmt"

	"github.com/vending-machine/server/internal/shared/valueobjects"
	"github.com/vending-machine/server/internal/transaction/domain"
)

// ConfirmSessionCommand is the input DTO for confirming a session
type ConfirmSessionCommand struct {
	SessionID  string
	PaymentRef string
}

// ConfirmSessionResult is the output DTO
type ConfirmSessionResult struct {
	SessionID  string
	TotalCents int64
	Currency   string
	PaymentRef string
}

// ConfirmSessionHandler orchestrates the session confirmation use case
type ConfirmSessionHandler struct {
	sessions  domain.SessionRepository
	publisher eventPublisher
}

func NewConfirmSessionHandler(sessions domain.SessionRepository, publisher eventPublisher) *ConfirmSessionHandler {
	if sessions == nil {
		panic("nil SessionRepository")
	}
	if publisher == nil {
		panic("nil EventPublisher")
	}
	return &ConfirmSessionHandler{
		sessions:  sessions,
		publisher: publisher,
	}
}

func (h *ConfirmSessionHandler) Handle(ctx context.Context, cmd ConfirmSessionCommand) (ConfirmSessionResult, error) {
	sessionID, err := valueobjects.SessionIDFrom(cmd.SessionID)
	if err != nil {
		return ConfirmSessionResult{}, fmt.Errorf("invalid session ID: %w", err)
	}

	sess, err := h.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return ConfirmSessionResult{}, domain.ErrSessionNotFound
	}

	if err := sess.Confirm(cmd.PaymentRef); err != nil {
		return ConfirmSessionResult{}, err
	}

	if err := h.sessions.Save(ctx, sess); err != nil {
		return ConfirmSessionResult{}, fmt.Errorf("failed to save session: %w", err)
	}

	// Publish domain events
	for _, evt := range sess.PullEvents() {
		_ = h.publisher.Publish(ctx, evt)
	}

	return ConfirmSessionResult{
		SessionID:  sess.ID().String(),
		TotalCents: sess.TotalAmount().Amount(),
		Currency:   sess.TotalAmount().Currency(),
		PaymentRef: cmd.PaymentRef,
	}, nil
}
