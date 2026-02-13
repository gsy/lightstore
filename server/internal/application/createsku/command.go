package createsku

// CreateSKUCommand is the input DTO for creating a SKU
type CreateSKUCommand struct {
	Code            string
	Name            string
	PriceCents      int64
	Currency        string
	WeightGrams     float64
	WeightTolerance float64
	ImageURL        string
}
