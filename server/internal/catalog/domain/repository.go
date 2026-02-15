package domain

import (
	"context"

	"github.com/vending-machine/server/internal/shared/valueobjects"
)

// SKURepository is the PORT interface defined by the domain
type SKURepository interface {
	Save(ctx context.Context, sku *SKU) error
	FindByID(ctx context.Context, id valueobjects.SKUID) (*SKU, error)
	FindByCode(ctx context.Context, code string) (*SKU, error)
	FindAllActive(ctx context.Context) ([]*SKU, error)
	FindAll(ctx context.Context) ([]*SKU, error)
}
