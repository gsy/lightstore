package support

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestContext holds shared state between BDD steps
type TestContext struct {
	// Server
	Server *httptest.Server
	Client *http.Client

	// Database
	DBPool *pgxpool.Pool

	// Request/Response state
	LastRequest  *http.Request
	LastResponse *http.Response
	LastBody     []byte

	// Test data storage
	CreatedSKUs     map[string]string // code -> id
	CreatedDevices  map[string]string // machine_id -> id
	CreatedSessions map[string]string // label -> session_id
}

// NewTestContext creates a new test context
func NewTestContext() *TestContext {
	return &TestContext{
		Client:          &http.Client{},
		CreatedSKUs:     make(map[string]string),
		CreatedDevices:  make(map[string]string),
		CreatedSessions: make(map[string]string),
	}
}

// SendRequest sends an HTTP request and stores the response
func (tc *TestContext) SendRequest(method, path string, body interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := tc.Server.URL + path
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	tc.LastRequest = req

	resp, err := tc.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	tc.LastResponse = resp
	tc.LastBody, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	return nil
}

// GetResponseJSON unmarshals the last response body into a map
func (tc *TestContext) GetResponseJSON() (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(tc.LastBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return result, nil
}

// GetNestedField retrieves a nested field from the response using dot notation
// e.g., "session.status" returns response["session"]["status"]
func (tc *TestContext) GetNestedField(path string) (interface{}, error) {
	response, err := tc.GetResponseJSON()
	if err != nil {
		return nil, err
	}

	parts := strings.Split(path, ".")
	var current interface{} = response

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil, fmt.Errorf("field %s not found in path %s", part, path)
		}
	}

	return current, nil
}

// CleanDatabase truncates all tables
func (tc *TestContext) CleanDatabase() error {
	ctx := context.Background()

	tables := []string{
		"transaction_detected_items",
		"transaction_sessions",
		"catalog_skus",
		"device_devices",
	}

	for _, table := range tables {
		_, err := tc.DBPool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}

	// Clear test data maps
	tc.CreatedSKUs = make(map[string]string)
	tc.CreatedDevices = make(map[string]string)
	tc.CreatedSessions = make(map[string]string)

	return nil
}

// Reset clears request/response state
func (tc *TestContext) Reset() {
	tc.LastRequest = nil
	tc.LastResponse = nil
	tc.LastBody = nil
}
