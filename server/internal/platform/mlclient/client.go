// Package mlclient provides a gRPC client for the ML detection server.
package mlclient

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/gsy/lightstore/server/internal/platform/mlclient/generated"
)

// Detection represents a single detected object.
type Detection struct {
	ClassName   string
	SKUID       string
	ClassID     int32
	Confidence  float32
	BoundingBox BoundingBox
}

// BoundingBox represents normalized bounding box coordinates.
type BoundingBox struct {
	X1, Y1, X2, Y2 float32
}

// DetectResult contains the detection response.
type DetectResult struct {
	Detections      []Detection
	ModelVersion    string
	InferenceTimeMs float32
	RequestID       string
}

// ModelInfo contains model metadata.
type ModelInfo struct {
	Version      string
	Architecture string
	ClassNames   []string
	InputWidth   int32
	InputHeight  int32
	MAP50        float32
	MAP50_95     float32
}

// HealthStatus contains server health information.
type HealthStatus struct {
	Healthy       bool
	Status        string
	ModelLoaded   bool
	UptimeSeconds int64
}

// Client is a gRPC client for the ML detection server.
type Client struct {
	conn   *grpc.ClientConn
	client pb.DetectionServiceClient
}

// Config holds client configuration.
type Config struct {
	Address     string
	DialTimeout time.Duration
}

// DefaultConfig returns default client configuration.
func DefaultConfig() Config {
	return Config{
		Address:     "localhost:50051",
		DialTimeout: 10 * time.Second,
	}
}

// New creates a new ML client.
func New(cfg Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ML server: %w", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewDetectionServiceClient(conn),
	}, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// DetectOptions holds options for detection request.
type DetectOptions struct {
	DeviceID            string
	ConfidenceThreshold float32
	IoUThreshold        float32
}

// Detect performs object detection on an image.
func (c *Client) Detect(ctx context.Context, imageBytes []byte, opts DetectOptions) (*DetectResult, error) {
	req := &pb.DetectRequest{
		Image:               imageBytes,
		DeviceId:            opts.DeviceID,
		ConfidenceThreshold: opts.ConfidenceThreshold,
		IouThreshold:        opts.IoUThreshold,
	}

	resp, err := c.client.Detect(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("detection failed: %w", err)
	}

	result := &DetectResult{
		ModelVersion:    resp.ModelVersion,
		InferenceTimeMs: resp.InferenceTimeMs,
		RequestID:       resp.RequestId,
		Detections:      make([]Detection, 0, len(resp.Detections)),
	}

	for _, d := range resp.Detections {
		det := Detection{
			ClassName:  d.ClassName,
			SKUID:      d.SkuId,
			ClassID:    d.ClassId,
			Confidence: d.Confidence,
		}
		if d.Bbox != nil {
			det.BoundingBox = BoundingBox{
				X1: d.Bbox.X1,
				Y1: d.Bbox.Y1,
				X2: d.Bbox.X2,
				Y2: d.Bbox.Y2,
			}
		}
		result.Detections = append(result.Detections, det)
	}

	return result, nil
}

// HealthCheck checks the ML server health.
func (c *Client) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	resp, err := c.client.HealthCheck(ctx, &pb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	return &HealthStatus{
		Healthy:       resp.Healthy,
		Status:        resp.Status,
		ModelLoaded:   resp.ModelLoaded,
		UptimeSeconds: resp.UptimeSeconds,
	}, nil
}

// GetModelInfo retrieves model metadata.
func (c *Client) GetModelInfo(ctx context.Context) (*ModelInfo, error) {
	resp, err := c.client.GetModelInfo(ctx, &pb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("get model info failed: %w", err)
	}

	return &ModelInfo{
		Version:      resp.Version,
		Architecture: resp.Architecture,
		ClassNames:   resp.ClassNames,
		InputWidth:   resp.InputWidth,
		InputHeight:  resp.InputHeight,
		MAP50:        resp.MAP50,
		MAP50_95:     resp.MAP50_95,
	}, nil
}

// ClassMapping represents a class ID to SKU mapping.
type ClassMapping struct {
	ClassID   int32
	SKUID     string
	ClassName string
}

// SyncClasses updates the ML server's class mapping.
func (c *Client) SyncClasses(ctx context.Context, mappings []ClassMapping) (int32, error) {
	classes := make([]*pb.ClassMapping, 0, len(mappings))
	for _, m := range mappings {
		classes = append(classes, &pb.ClassMapping{
			ClassId:   m.ClassID,
			SkuId:     m.SKUID,
			ClassName: m.ClassName,
		})
	}

	req := &pb.SyncClassesRequest{Classes: classes}
	resp, err := c.client.SyncClasses(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("sync classes failed: %w", err)
	}

	if !resp.Success {
		return 0, fmt.Errorf("sync classes returned failure")
	}

	return resp.ClassCount, nil
}
