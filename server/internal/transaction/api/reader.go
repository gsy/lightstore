package api

import (
	"context"

	"github.com/vending-machine/server/internal/transaction/app"
)

// SessionView is the DTO exposed to other contexts
type SessionView struct {
	ID          string
	DeviceID    string
	UserID      string
	Status      string
	TotalCents  int64
	Currency    string
	TotalWeight float64
}

// SessionReader is the interface exposed to other contexts for reading session data
type SessionReader interface {
	FindByID(ctx context.Context, id string) (*SessionView, error)
	FindActiveByDeviceID(ctx context.Context, deviceID string) (*SessionView, error)
}

// SessionReaderAdapter implements SessionReader using the app layer query service
type SessionReaderAdapter struct {
	queryService *app.SessionQueryService
}

func NewSessionReaderAdapter(queryService *app.SessionQueryService) *SessionReaderAdapter {
	return &SessionReaderAdapter{queryService: queryService}
}

func (a *SessionReaderAdapter) FindByID(ctx context.Context, id string) (*SessionView, error) {
	view, err := a.queryService.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &SessionView{
		ID:          view.ID,
		DeviceID:    view.DeviceID,
		UserID:      view.UserID,
		Status:      view.Status,
		TotalCents:  view.TotalCents,
		Currency:    view.Currency,
		TotalWeight: view.TotalWeight,
	}, nil
}

func (a *SessionReaderAdapter) FindActiveByDeviceID(ctx context.Context, deviceID string) (*SessionView, error) {
	view, err := a.queryService.FindActiveByDeviceID(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	return &SessionView{
		ID:          view.ID,
		DeviceID:    view.DeviceID,
		UserID:      view.UserID,
		Status:      view.Status,
		TotalCents:  view.TotalCents,
		Currency:    view.Currency,
		TotalWeight: view.TotalWeight,
	}, nil
}
