package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vending-machine/server/internal/domain/session"
	"github.com/vending-machine/server/internal/domain/shared"
)

// PostgresSessionRepository implements session.SessionRepository
type PostgresSessionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSessionRepository(pool *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{pool: pool}
}

type sessionRow struct {
	ID           string
	DeviceID     string
	UserID       *string
	Status       string
	Items        []byte
	TotalWeight  float64
	TotalCents   int64
	Currency     string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	CompletedAt  *time.Time
}

type itemJSON struct {
	SKUID      string  `json:"sku_id"`
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	PriceCents int64   `json:"price_cents"`
	Currency   string  `json:"currency"`
}

func (r *PostgresSessionRepository) Save(ctx context.Context, s *session.Session) error {
	var userID *string
	if s.UserID() != "" {
		u := s.UserID()
		userID = &u
	}

	// Serialize detected items
	var itemsJSON []itemJSON
	for _, item := range s.DetectedItems() {
		itemsJSON = append(itemsJSON, itemJSON{
			SKUID:      item.SKUID().String(),
			Code:       item.Code(),
			Name:       item.Name(),
			Confidence: item.Confidence(),
			PriceCents: item.Price().Amount(),
			Currency:   item.Price().Currency(),
		})
	}
	itemsData, _ := json.Marshal(itemsJSON)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (id, device_id, user_id, status, items, total_weight, total_cents, currency, created_at, expires_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			items = EXCLUDED.items,
			total_weight = EXCLUDED.total_weight,
			total_cents = EXCLUDED.total_cents,
			currency = EXCLUDED.currency,
			completed_at = EXCLUDED.completed_at
	`, s.ID().String(), s.DeviceID().String(), userID, string(s.Status()),
		itemsData, s.TotalWeight().Grams(), s.TotalAmount().Amount(), s.TotalAmount().Currency(),
		s.CreatedAt(), s.ExpiresAt(), s.CompletedAt())

	return err
}

func (r *PostgresSessionRepository) FindByID(ctx context.Context, id shared.SessionID) (*session.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, device_id, user_id, status, items, total_weight, total_cents, currency, created_at, expires_at, completed_at
		FROM sessions WHERE id = $1
	`, id.String())

	return r.scanSession(row)
}

func (r *PostgresSessionRepository) FindActiveByDeviceID(ctx context.Context, deviceID shared.DeviceID) (*session.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, device_id, user_id, status, items, total_weight, total_cents, currency, created_at, expires_at, completed_at
		FROM sessions
		WHERE device_id = $1 AND status = 'active' AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`, deviceID.String())

	return r.scanSession(row)
}

func (r *PostgresSessionRepository) scanSession(row pgx.Row) (*session.Session, error) {
	var rec sessionRow
	err := row.Scan(
		&rec.ID, &rec.DeviceID, &rec.UserID, &rec.Status, &rec.Items,
		&rec.TotalWeight, &rec.TotalCents, &rec.Currency,
		&rec.CreatedAt, &rec.ExpiresAt, &rec.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, session.ErrSessionNotFound
		}
		return nil, err
	}

	return r.reconstitute(rec), nil
}

func (r *PostgresSessionRepository) reconstitute(rec sessionRow) *session.Session {
	id, _ := shared.SessionIDFrom(rec.ID)
	deviceID, _ := shared.DeviceIDFrom(rec.DeviceID)

	userID := ""
	if rec.UserID != nil {
		userID = *rec.UserID
	}

	// Parse items
	var itemsJSON []itemJSON
	_ = json.Unmarshal(rec.Items, &itemsJSON)

	var detectedItems []session.DetectedItem
	for _, item := range itemsJSON {
		skuID, _ := shared.SKUIDFrom(item.SKUID)
		price, _ := shared.NewMoney(item.PriceCents, item.Currency)
		detectedItems = append(detectedItems, session.NewDetectedItem(
			skuID,
			item.Code,
			item.Name,
			item.Confidence,
			price,
		))
	}

	totalWeight, _ := shared.NewWeight(rec.TotalWeight)
	totalAmount, _ := shared.NewMoney(rec.TotalCents, rec.Currency)

	return session.Reconstitute(
		id,
		deviceID,
		userID,
		session.SessionStatus(rec.Status),
		detectedItems,
		totalWeight,
		totalAmount,
		rec.CreatedAt,
		rec.ExpiresAt,
		rec.CompletedAt,
	)
}
