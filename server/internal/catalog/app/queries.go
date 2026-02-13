package app

import (
	"context"

	"github.com/vending-machine/server/internal/catalog/domain"
	"github.com/vending-machine/server/internal/shared/valueobjects"
)

// SKUQueryService provides read-only access to SKUs for the catalog context's HTTP layer
type SKUQueryService struct {
	repo domain.SKURepository
}

func NewSKUQueryService(repo domain.SKURepository) *SKUQueryService {
	return &SKUQueryService{repo: repo}
}

func (s *SKUQueryService) FindByID(ctx context.Context, id string) (*domain.SKU, error) {
	skuID, err := valueobjects.SKUIDFrom(id)
	if err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, skuID)
}

func (s *SKUQueryService) FindAll(ctx context.Context) ([]*domain.SKU, error) {
	return s.repo.FindAll(ctx)
}

func (s *SKUQueryService) FindAllActive(ctx context.Context) ([]*domain.SKU, error) {
	return s.repo.FindAllActive(ctx)
}
