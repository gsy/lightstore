package api

import (
	"context"

	"github.com/vending-machine/server/internal/catalog/domain"
	"github.com/vending-machine/server/internal/shared/valueobjects"
)

// SKUView is a read-only DTO exposed to other bounded contexts
type SKUView struct {
	ID              string
	Code            string
	Name            string
	PriceCents      int64
	Currency        string
	WeightGrams     float64
	WeightTolerance float64
	ImageURL        string
	Active          bool
}

// SKUReader is the interface other contexts use to read catalog data.
// This prevents direct domain coupling between bounded contexts.
type SKUReader interface {
	FindByCode(ctx context.Context, code string) (*SKUView, error)
	FindByID(ctx context.Context, id string) (*SKUView, error)
	FindAllActive(ctx context.Context) ([]SKUView, error)
	FindAll(ctx context.Context) ([]SKUView, error)
}

// SKUReaderAdapter implements SKUReader using the domain repository
type SKUReaderAdapter struct {
	repo domain.SKURepository
}

func NewSKUReaderAdapter(repo domain.SKURepository) *SKUReaderAdapter {
	return &SKUReaderAdapter{repo: repo}
}

func (a *SKUReaderAdapter) FindByCode(ctx context.Context, code string) (*SKUView, error) {
	sku, err := a.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return toSKUView(sku), nil
}

func (a *SKUReaderAdapter) FindByID(ctx context.Context, id string) (*SKUView, error) {
	skuID, err := valueobjects.SKUIDFrom(id)
	if err != nil {
		return nil, err
	}
	sku, err := a.repo.FindByID(ctx, skuID)
	if err != nil {
		return nil, err
	}
	return toSKUView(sku), nil
}

func (a *SKUReaderAdapter) FindAllActive(ctx context.Context) ([]SKUView, error) {
	skus, err := a.repo.FindAllActive(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]SKUView, len(skus))
	for i, sku := range skus {
		views[i] = *toSKUView(sku)
	}
	return views, nil
}

func (a *SKUReaderAdapter) FindAll(ctx context.Context) ([]SKUView, error) {
	skus, err := a.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]SKUView, len(skus))
	for i, sku := range skus {
		views[i] = *toSKUView(sku)
	}
	return views, nil
}

func toSKUView(sku *domain.SKU) *SKUView {
	return &SKUView{
		ID:              sku.ID().String(),
		Code:            sku.Code(),
		Name:            sku.Name(),
		PriceCents:      sku.Price().Amount(),
		Currency:        sku.Price().Currency(),
		WeightGrams:     sku.Weight().Grams(),
		WeightTolerance: sku.WeightTolerance(),
		ImageURL:        sku.ImageURL(),
		Active:          sku.IsActive(),
	}
}
