package domain

import "github.com/vending-machine/server/internal/shared/valueobjects"

// DetectedItem is a value object representing a detected SKU
type DetectedItem struct {
	skuID      valueobjects.SKUID
	code       string
	name       string
	confidence float64
	price      valueobjects.Money
}

func NewDetectedItem(skuID valueobjects.SKUID, code, name string, confidence float64, price valueobjects.Money) DetectedItem {
	return DetectedItem{
		skuID:      skuID,
		code:       code,
		name:       name,
		confidence: confidence,
		price:      price,
	}
}

func (d DetectedItem) SKUID() valueobjects.SKUID  { return d.skuID }
func (d DetectedItem) Code() string               { return d.code }
func (d DetectedItem) Name() string               { return d.name }
func (d DetectedItem) Confidence() float64        { return d.confidence }
func (d DetectedItem) Price() valueobjects.Money  { return d.price }
