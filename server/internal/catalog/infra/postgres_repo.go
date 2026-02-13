package infra

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vending-machine/server/internal/catalog/domain"
	"github.com/vending-machine/server/internal/shared/valueobjects"
)

// PostgresSKURepository implements domain.SKURepository
type PostgresSKURepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSKURepository(pool *pgxpool.Pool) *PostgresSKURepository {
	return &PostgresSKURepository{pool: pool}
}

// skuRow is a DB-layer struct (never leaves this file)
type skuRow struct {
	ID              string
	Code            string
	Name            string
	PriceCents      int64
	Currency        string
	WeightGrams     float64
	WeightTolerance float64
	ImageURL        *string
	Active          bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (r *PostgresSKURepository) Save(ctx context.Context, s *domain.SKU) error {
	var imageURL *string
	if s.ImageURL() != "" {
		url := s.ImageURL()
		imageURL = &url
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO skus (id, code, name, price_cents, currency, weight_grams, weight_tolerance, image_url, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			price_cents = EXCLUDED.price_cents,
			currency = EXCLUDED.currency,
			weight_grams = EXCLUDED.weight_grams,
			weight_tolerance = EXCLUDED.weight_tolerance,
			image_url = EXCLUDED.image_url,
			active = EXCLUDED.active,
			updated_at = EXCLUDED.updated_at
	`, s.ID().String(), s.Code(), s.Name(), s.Price().Amount(), s.Price().Currency(),
		s.Weight().Grams(), s.WeightTolerance(), imageURL, s.IsActive(), s.CreatedAt(), s.UpdatedAt())

	return err
}

func (r *PostgresSKURepository) FindByID(ctx context.Context, id valueobjects.SKUID) (*domain.SKU, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, code, name, price_cents, currency, weight_grams, weight_tolerance, image_url, active, created_at, updated_at
		FROM skus WHERE id = $1
	`, id.String())

	return r.scanSKU(row)
}

func (r *PostgresSKURepository) FindByCode(ctx context.Context, code string) (*domain.SKU, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, code, name, price_cents, currency, weight_grams, weight_tolerance, image_url, active, created_at, updated_at
		FROM skus WHERE code = $1
	`, code)

	return r.scanSKU(row)
}

func (r *PostgresSKURepository) FindAllActive(ctx context.Context) ([]*domain.SKU, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, price_cents, currency, weight_grams, weight_tolerance, image_url, active, created_at, updated_at
		FROM skus WHERE active = true ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSKUs(rows)
}

func (r *PostgresSKURepository) FindAll(ctx context.Context) ([]*domain.SKU, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name, price_cents, currency, weight_grams, weight_tolerance, image_url, active, created_at, updated_at
		FROM skus ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanSKUs(rows)
}

func (r *PostgresSKURepository) scanSKU(row pgx.Row) (*domain.SKU, error) {
	var rec skuRow
	err := row.Scan(
		&rec.ID, &rec.Code, &rec.Name, &rec.PriceCents, &rec.Currency,
		&rec.WeightGrams, &rec.WeightTolerance, &rec.ImageURL, &rec.Active,
		&rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSKUNotFound
		}
		return nil, err
	}

	return r.reconstitute(rec), nil
}

func (r *PostgresSKURepository) scanSKUs(rows pgx.Rows) ([]*domain.SKU, error) {
	var skus []*domain.SKU
	for rows.Next() {
		var rec skuRow
		err := rows.Scan(
			&rec.ID, &rec.Code, &rec.Name, &rec.PriceCents, &rec.Currency,
			&rec.WeightGrams, &rec.WeightTolerance, &rec.ImageURL, &rec.Active,
			&rec.CreatedAt, &rec.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		skus = append(skus, r.reconstitute(rec))
	}
	return skus, nil
}

func (r *PostgresSKURepository) reconstitute(rec skuRow) *domain.SKU {
	id, _ := valueobjects.SKUIDFrom(rec.ID)
	price, _ := valueobjects.NewMoney(rec.PriceCents, rec.Currency)
	weight, _ := valueobjects.NewWeight(rec.WeightGrams)

	imageURL := ""
	if rec.ImageURL != nil {
		imageURL = *rec.ImageURL
	}

	return domain.Reconstitute(
		id,
		rec.Code,
		rec.Name,
		price,
		weight,
		rec.WeightTolerance,
		imageURL,
		rec.Active,
		rec.CreatedAt,
		rec.UpdatedAt,
	)
}
