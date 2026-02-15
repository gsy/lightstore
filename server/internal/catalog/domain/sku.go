package domain

import (
	"time"

	"github.com/vending-machine/server/internal/shared/events"
	"github.com/vending-machine/server/internal/shared/valueobjects"
)

// SKU is the aggregate root for product management
type SKU struct {
	id              valueobjects.SKUID
	code            string // unique identifier code
	name            string
	price           valueobjects.Money
	weight          valueobjects.Weight
	weightTolerance float64
	imageURL        string
	active          bool
	createdAt       time.Time
	updatedAt       time.Time

	domainEvents []events.DomainEvent
}

// NewSKU creates a new SKU with validation
func NewSKU(code, name string, priceCents int64, currency string, weightGrams float64) (*SKU, error) {
	if code == "" {
		return nil, ErrInvalidSKUCode
	}
	if name == "" {
		return nil, ErrInvalidSKUName
	}

	price, err := valueobjects.NewMoney(priceCents, currency)
	if err != nil {
		return nil, ErrInvalidSKUPrice
	}

	weight, err := valueobjects.NewWeight(weightGrams)
	if err != nil {
		return nil, ErrInvalidSKUWeight
	}

	now := time.Now().UTC()
	s := &SKU{
		id:              valueobjects.NewSKUID(),
		code:            code,
		name:            name,
		price:           price,
		weight:          weight,
		weightTolerance: 5.0, // default tolerance in grams
		active:          true,
		createdAt:       now,
		updatedAt:       now,
	}

	s.domainEvents = append(s.domainEvents, NewSKUCreated(s.id, code, name))

	return s, nil
}

// Reconstitute rebuilds a SKU from persistence (no validation, no events)
func Reconstitute(
	id valueobjects.SKUID,
	code, name string,
	price valueobjects.Money,
	weight valueobjects.Weight,
	weightTolerance float64,
	imageURL string,
	active bool,
	createdAt, updatedAt time.Time,
) *SKU {
	return &SKU{
		id:              id,
		code:            code,
		name:            name,
		price:           price,
		weight:          weight,
		weightTolerance: weightTolerance,
		imageURL:        imageURL,
		active:          active,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

// Getters
func (s *SKU) ID() valueobjects.SKUID      { return s.id }
func (s *SKU) Code() string                { return s.code }
func (s *SKU) Name() string                { return s.name }
func (s *SKU) Price() valueobjects.Money   { return s.price }
func (s *SKU) Weight() valueobjects.Weight { return s.weight }
func (s *SKU) WeightTolerance() float64    { return s.weightTolerance }
func (s *SKU) ImageURL() string            { return s.imageURL }
func (s *SKU) IsActive() bool              { return s.active }
func (s *SKU) CreatedAt() time.Time        { return s.createdAt }
func (s *SKU) UpdatedAt() time.Time        { return s.updatedAt }

// Business methods

func (s *SKU) Update(name string, priceCents int64, currency string, weightGrams, weightTolerance float64, imageURL string) error {
	if name == "" {
		return ErrInvalidSKUName
	}

	price, err := valueobjects.NewMoney(priceCents, currency)
	if err != nil {
		return ErrInvalidSKUPrice
	}

	weight, err := valueobjects.NewWeight(weightGrams)
	if err != nil {
		return ErrInvalidSKUWeight
	}

	s.name = name
	s.price = price
	s.weight = weight
	s.weightTolerance = weightTolerance
	s.imageURL = imageURL
	s.updatedAt = time.Now().UTC()

	s.domainEvents = append(s.domainEvents, NewSKUUpdated(s.id, name))

	return nil
}

func (s *SKU) Deactivate() {
	if !s.active {
		return
	}
	s.active = false
	s.updatedAt = time.Now().UTC()
	s.domainEvents = append(s.domainEvents, NewSKUDeactivated(s.id))
}

func (s *SKU) Activate() {
	if s.active {
		return
	}
	s.active = true
	s.updatedAt = time.Now().UTC()
}

func (s *SKU) IsWeightMatch(measured valueobjects.Weight) bool {
	return s.weight.IsWithinTolerance(measured, s.weightTolerance)
}

// PullEvents returns accumulated domain events and clears the slice
func (s *SKU) PullEvents() []events.DomainEvent {
	evts := s.domainEvents
	s.domainEvents = nil
	return evts
}
