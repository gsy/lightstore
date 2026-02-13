package submitdetection

// DetectedItemInput represents a detected item from the device
type DetectedItemInput struct {
	SKU        string
	Confidence float64
	BBox       []float64
}

// SubmitDetectionCommand is the input DTO for submitting detection results
type SubmitDetectionCommand struct {
	DeviceID    string
	SessionID   string
	Items       []DetectedItemInput
	TotalWeight float64
}
