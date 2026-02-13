package ports

import "context"

// MLDetectionResult represents the result from ML inference
type MLDetectionResult struct {
	SKU        string
	Confidence float64
	BBox       []float64
}

// MLService is an OUTPUT PORT for ML inference
type MLService interface {
	// DetectFromImage performs cloud ML inference on an image
	DetectFromImage(ctx context.Context, imageData []byte) ([]MLDetectionResult, error)
}
