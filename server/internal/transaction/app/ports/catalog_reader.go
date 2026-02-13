package ports

import "context"

// SKUInfo is a DTO representing SKU information needed by transaction context
type SKUInfo struct {
	ID          string
	Code        string
	Name        string
	PriceCents  int64
	Currency    string
	WeightGrams float64
}

// CatalogReader is an input port for reading catalog context data.
// This port is defined by the transaction context (consumer) and
// implemented by an adapter that calls the catalog context API.
type CatalogReader interface {
	FindSKUByCode(ctx context.Context, code string) (*SKUInfo, error)
}
