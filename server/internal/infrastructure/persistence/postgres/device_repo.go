package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vending-machine/server/internal/domain/device"
	"github.com/vending-machine/server/internal/domain/shared"
)

// PostgresDeviceRepository implements device.DeviceRepository
type PostgresDeviceRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDeviceRepository(pool *pgxpool.Pool) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{pool: pool}
}

type deviceRow struct {
	ID        string
	MachineID string
	Name      *string
	Location  *string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (r *PostgresDeviceRepository) Save(ctx context.Context, d *device.Device) error {
	var name, location *string
	if d.Name() != "" {
		n := d.Name()
		name = &n
	}
	if d.Location() != "" {
		l := d.Location()
		location = &l
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO devices (id, machine_id, name, location, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			location = EXCLUDED.location,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`, d.ID().String(), d.MachineID(), name, location, string(d.Status()), d.CreatedAt(), d.UpdatedAt())

	return err
}

func (r *PostgresDeviceRepository) FindByID(ctx context.Context, id shared.DeviceID) (*device.Device, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, machine_id, name, location, status, created_at, updated_at
		FROM devices WHERE id = $1
	`, id.String())

	return r.scanDevice(row)
}

func (r *PostgresDeviceRepository) FindByMachineID(ctx context.Context, machineID string) (*device.Device, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, machine_id, name, location, status, created_at, updated_at
		FROM devices WHERE machine_id = $1
	`, machineID)

	return r.scanDevice(row)
}

func (r *PostgresDeviceRepository) scanDevice(row pgx.Row) (*device.Device, error) {
	var rec deviceRow
	err := row.Scan(
		&rec.ID, &rec.MachineID, &rec.Name, &rec.Location,
		&rec.Status, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, device.ErrDeviceNotFound
		}
		return nil, err
	}

	return r.reconstitute(rec), nil
}

func (r *PostgresDeviceRepository) reconstitute(rec deviceRow) *device.Device {
	id, _ := shared.DeviceIDFrom(rec.ID)

	name := ""
	if rec.Name != nil {
		name = *rec.Name
	}
	location := ""
	if rec.Location != nil {
		location = *rec.Location
	}

	return device.Reconstitute(
		id,
		rec.MachineID,
		name,
		location,
		device.DeviceStatus(rec.Status),
		rec.CreatedAt,
		rec.UpdatedAt,
	)
}
