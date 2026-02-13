package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vending-machine/server/internal/pkg/logger"
)

// RunMigrations executes all database migrations for all bounded contexts
func RunMigrations(pool *pgxpool.Pool) error {
	migrations := []string{
		// =========================================================================
		// Catalog Context Tables
		// =========================================================================
		`CREATE TABLE IF NOT EXISTS skus (
			id UUID PRIMARY KEY,
			code VARCHAR(50) UNIQUE NOT NULL,
			name VARCHAR(100) NOT NULL,
			price_cents BIGINT NOT NULL,
			currency VARCHAR(3) NOT NULL DEFAULT 'USD',
			weight_grams DECIMAL(10,1) NOT NULL,
			weight_tolerance DECIMAL(10,1) DEFAULT 5.0,
			image_url VARCHAR(500),
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// =========================================================================
		// Device Context Tables
		// =========================================================================
		`CREATE TABLE IF NOT EXISTS devices (
			id UUID PRIMARY KEY,
			machine_id VARCHAR(50) UNIQUE NOT NULL,
			name VARCHAR(100),
			location VARCHAR(200),
			status VARCHAR(20) DEFAULT 'active',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// =========================================================================
		// Transaction Context Tables
		// =========================================================================
		`CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY,
			device_id UUID REFERENCES devices(id),
			user_id VARCHAR(100),
			status VARCHAR(20) DEFAULT 'active',
			items JSONB DEFAULT '[]',
			total_weight DECIMAL(10,1) DEFAULT 0,
			total_cents BIGINT DEFAULT 0,
			currency VARCHAR(3) DEFAULT 'USD',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			expires_at TIMESTAMP WITH TIME ZONE,
			completed_at TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY,
			session_id UUID REFERENCES sessions(id),
			items JSONB NOT NULL,
			total_cents BIGINT NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD',
			status VARCHAR(20) DEFAULT 'pending',
			payment_ref VARCHAR(100),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			completed_at TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE TABLE IF NOT EXISTS refunds (
			id UUID PRIMARY KEY,
			transaction_id UUID REFERENCES transactions(id),
			reason TEXT,
			amount_cents BIGINT NOT NULL,
			currency VARCHAR(3) DEFAULT 'USD',
			status VARCHAR(20) DEFAULT 'pending',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			processed_at TIMESTAMP WITH TIME ZONE
		)`,

		// =========================================================================
		// Indexes
		// =========================================================================
		`CREATE INDEX IF NOT EXISTS idx_sessions_device_id ON sessions(device_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_skus_code ON skus(code)`,
		`CREATE INDEX IF NOT EXISTS idx_skus_active ON skus(active)`,
	}

	for i, migration := range migrations {
		_, err := pool.Exec(context.Background(), migration)
		if err != nil {
			logger.Error("Migration failed", "migration", i, "error", err)
			return err
		}
	}

	logger.Info("Migrations completed", "count", len(migrations))
	return nil
}
