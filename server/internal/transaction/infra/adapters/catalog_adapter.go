package adapters

import (
	"context"

	catalogapi "github.com/vending-machine/server/internal/catalog/api"
	"github.com/vending-machine/server/internal/transaction/app/ports"
)

// CatalogAdapter implements ports.CatalogReader using the catalog context API
type CatalogAdapter struct {
	reader catalogapi.SKUReader
}

func NewCatalogAdapter(reader catalogapi.SKUReader) *CatalogAdapter {
	if reader == nil {
		panic("nil SKUReader")
	}
	return &CatalogAdapter{reader: reader}
}

func (a *CatalogAdapter) FindSKUByCode(ctx context.Context, code string) (*ports.SKUInfo, error) {
	view, err := a.reader.FindByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return &ports.SKUInfo{
		ID:          view.ID,
		Code:        view.Code,
		Name:        view.Name,
		PriceCents:  view.PriceCents,
		Currency:    view.Currency,
		WeightGrams: view.WeightGrams,
	}, nil
}
