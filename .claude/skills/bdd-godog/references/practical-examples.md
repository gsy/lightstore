# Godog Practical Examples

This document provides complete, runnable examples for common BDD testing scenarios using Godog.

---

## Example 1: REST API Testing

### Feature File

```gherkin
# features/api/catalog.feature
@api @catalog
Feature: Catalog API
  As a catalog manager
  I want to manage SKUs via REST API
  So that I can maintain product inventory

  Background:
    Given the API is running
    And I am authenticated with role "catalog_manager"

  @smoke
  Scenario: Create a new SKU
    When I send a POST request to "/api/v1/skus" with body:
      """
      {
        "code": "APPLE-001",
        "name": "Fuji Apple",
        "price": 2.50,
        "weight_grams": 150
      }
      """
    Then the response status should be 201
    And the response should contain:
      """
      {
        "id": "<any-uuid>",
        "code": "APPLE-001",
        "name": "Fuji Apple"
      }
      """

  Scenario: List all SKUs
    Given the following SKUs exist:
      | code      | name        | price | weight_grams |
      | APPLE-001 | Fuji Apple  | 2.50  | 150          |
      | APPLE-002 | Gala Apple  | 2.30  | 140          |
    When I send a GET request to "/api/v1/skus"
    Then the response status should be 200
    And the response should contain 2 SKUs

  Scenario: Get SKU by code
    Given a SKU exists with code "APPLE-001"
    When I send a GET request to "/api/v1/skus/APPLE-001"
    Then the response status should be 200
    And the response field "code" should be "APPLE-001"

  Scenario: Handle not found
    When I send a GET request to "/api/v1/skus/NONEXISTENT"
    Then the response status should be 404
```

### Step Definitions

```go
// test/bdd/api_steps_test.go
package bdd_test

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/cucumber/godog"
    "github.com/stretchr/testify/assert"
)

type apiTestContext struct {
    server       *httptest.Server
    client       *http.Client
    authToken    string
    lastRequest  *http.Request
    lastResponse *http.Response
    lastBody     []byte
    testData     map[string]interface{}
}

func TestAPIFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeAPIScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../../features/api"},
            TestingT: t,
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned")
    }
}

func InitializeAPIScenario(ctx *godog.ScenarioContext) {
    tc := &apiTestContext{
        client:   &http.Client{},
        testData: make(map[string]interface{}),
    }

    // Setup/teardown
    ctx.Before(tc.beforeScenario)
    ctx.After(tc.afterScenario)

    // Step definitions
    ctx.Step(`^the API is running$`, tc.apiIsRunning)
    ctx.Step(`^I am authenticated with role "([^"]*)"$`, tc.iAmAuthenticatedWithRole)
    ctx.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)"$`, tc.iSendRequestTo)
    ctx.Step(`^I send a (POST|PUT) request to "([^"]*)" with body:$`, tc.iSendRequestWithBody)
    ctx.Step(`^the response status should be (\d+)$`, tc.responseStatusShouldBe)
    ctx.Step(`^the response should contain:$`, tc.responseShouldContain)
    ctx.Step(`^the response should contain (\d+) SKUs$`, tc.responseShouldContainSKUs)
    ctx.Step(`^the response field "([^"]*)" should be "([^"]*)"$`, tc.responseFieldShouldBe)
    ctx.Step(`^a SKU exists with code "([^"]*)"$`, tc.skuExistsWithCode)
    ctx.Step(`^the following SKUs exist:$`, tc.followingSKUsExist)
}

func (tc *apiTestContext) beforeScenario(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
    // Start test server
    tc.server = startTestServer()
    
    // Clean database
    cleanDatabase()
    
    return ctx, nil
}

func (tc *apiTestContext) afterScenario(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
    if tc.server != nil {
        tc.server.Close()
    }
    return ctx, nil
}

func (tc *apiTestContext) apiIsRunning() error {
    resp, err := tc.client.Get(tc.server.URL + "/health")
    if err != nil {
        return fmt.Errorf("API not responding: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return fmt.Errorf("API health check failed: %d", resp.StatusCode)
    }
    
    return nil
}

func (tc *apiTestContext) iAmAuthenticatedWithRole(role string) error {
    // Generate test token
    token, err := generateTestToken(role)
    if err != nil {
        return err
    }
    tc.authToken = token
    return nil
}

func (tc *apiTestContext) iSendRequestTo(method, path string) error {
    return tc.sendRequest(method, path, nil)
}

func (tc *apiTestContext) iSendRequestWithBody(method, path string, body *godog.DocString) error {
    return tc.sendRequest(method, path, []byte(body.Content))
}

func (tc *apiTestContext) sendRequest(method, path string, body []byte) error {
    url := tc.server.URL + path
    
    var bodyReader io.Reader
    if body != nil {
        bodyReader = bytes.NewReader(body)
    }
    
    req, err := http.NewRequest(method, url, bodyReader)
    if err != nil {
        return err
    }
    
    if tc.authToken != "" {
        req.Header.Set("Authorization", "Bearer "+tc.authToken)
    }
    
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    
    tc.lastRequest = req
    
    resp, err := tc.client.Do(req)
    if err != nil {
        return err
    }
    
    tc.lastResponse = resp
    tc.lastBody, _ = io.ReadAll(resp.Body)
    resp.Body.Close()
    
    return nil
}

func (tc *apiTestContext) responseStatusShouldBe(expectedStatus int) error {
    if tc.lastResponse.StatusCode != expectedStatus {
        return fmt.Errorf("expected status %d, got %d. Body: %s",
            expectedStatus, tc.lastResponse.StatusCode, string(tc.lastBody))
    }
    return nil
}

func (tc *apiTestContext) responseShouldContain(expected *godog.DocString) error {
    var expectedJSON, actualJSON map[string]interface{}
    
    if err := json.Unmarshal([]byte(expected.Content), &expectedJSON); err != nil {
        return fmt.Errorf("invalid expected JSON: %w", err)
    }
    
    if err := json.Unmarshal(tc.lastBody, &actualJSON); err != nil {
        return fmt.Errorf("invalid response JSON: %w", err)
    }
    
    // Compare fields (ignoring <any-uuid> placeholders)
    for key, expectedValue := range expectedJSON {
        actualValue, exists := actualJSON[key]
        if !exists {
            return fmt.Errorf("field %s not found in response", key)
        }
        
        // Handle placeholders
        if strValue, ok := expectedValue.(string); ok && strings.HasPrefix(strValue, "<any-") {
            continue
        }
        
        if !assert.ObjectsAreEqual(expectedValue, actualValue) {
            return fmt.Errorf("field %s: expected %v, got %v", key, expectedValue, actualValue)
        }
    }
    
    return nil
}

func (tc *apiTestContext) responseShouldContainSKUs(count int) error {
    var response struct {
        SKUs []interface{} `json:"skus"`
    }
    
    if err := json.Unmarshal(tc.lastBody, &response); err != nil {
        return err
    }
    
    if len(response.SKUs) != count {
        return fmt.Errorf("expected %d SKUs, got %d", count, len(response.SKUs))
    }
    
    return nil
}

func (tc *apiTestContext) responseFieldShouldBe(field, expectedValue string) error {
    var response map[string]interface{}
    
    if err := json.Unmarshal(tc.lastBody, &response); err != nil {
        return err
    }
    
    actualValue, exists := response[field]
    if !exists {
        return fmt.Errorf("field %s not found", field)
    }
    
    if fmt.Sprint(actualValue) != expectedValue {
        return fmt.Errorf("field %s: expected %s, got %v", field, expectedValue, actualValue)
    }
    
    return nil
}

func (tc *apiTestContext) skuExistsWithCode(code string) error {
    sku := map[string]interface{}{
        "code":          code,
        "name":          "Test SKU",
        "price":         1.99,
        "weight_grams":  100,
    }
    
    body, _ := json.Marshal(sku)
    return tc.sendRequest("POST", "/api/v1/skus", body)
}

func (tc *apiTestContext) followingSKUsExist(table *godog.Table) error {
    for i, row := range table.Rows {
        if i == 0 {
            continue // Skip header
        }
        
        sku := map[string]interface{}{
            "code":          row.Cells[0].Value,
            "name":          row.Cells[1].Value,
            "price":         parseFloat(row.Cells[2].Value),
            "weight_grams":  parseInt(row.Cells[3].Value),
        }
        
        body, _ := json.Marshal(sku)
        if err := tc.sendRequest("POST", "/api/v1/skus", body); err != nil {
            return err
        }
    }
    
    return nil
}
```

---

## Example 2: Domain Logic Testing

### Feature File

```gherkin
# features/domain/session.feature
@domain @transaction
Feature: Shopping Session
  As a customer
  I want to manage my shopping session
  So that I can purchase items

  Background:
    Given a device exists with machine ID "DEVICE-001"
    And the following SKUs exist in catalog:
      | code      | name       | price | weight_grams |
      | APPLE-001 | Fuji Apple | 2.50  | 150          |
      | APPLE-002 | Gala Apple | 2.30  | 140          |

  Scenario: Start a new session
    When I start a session on device "DEVICE-001"
    Then the session should be in "active" state
    And the session should have 0 items
    And the total should be $0.00

  Scenario: Add items to session
    Given an active session exists on device "DEVICE-001"
    When I add the following detections:
      | sku_code  | weight_grams |
      | APPLE-001 | 150          |
      | APPLE-002 | 140          |
    Then the session should have 2 items
    And the total should be $4.80

  Scenario: Cannot add items to confirmed session
    Given a confirmed session exists on device "DEVICE-001"
    When I try to add a detection for "APPLE-001"
    Then I should receive an error "session already confirmed"
    And the session should remain unchanged

  Scenario: Confirm session
    Given an active session with items exists on device "DEVICE-001"
    When I confirm the session
    Then the session should be in "confirmed" state
    And a "SessionConfirmed" event should be published

  Scenario: Cancel session
    Given an active session exists on device "DEVICE-001"
    When I cancel the session
    Then the session should be in "cancelled" state
    And a "SessionCancelled" event should be published
```

### Step Definitions

```go
// internal/transaction/bdd_test.go
package transaction_test

import (
    "context"
    "fmt"
    "testing"

    "github.com/cucumber/godog"
    
    "myservice/internal/catalog/domain"
    "myservice/internal/device/domain"
    "myservice/internal/transaction/app"
    "myservice/internal/transaction/domain"
)

type transactionTestContext struct {
    // In-memory repositories
    deviceRepo      *InMemoryDeviceRepository
    skuRepo         *InMemorySKURepository
    sessionRepo     *InMemorySessionRepository
    eventBus        *InMemoryEventBus
    
    // Application services
    startSession    *app.StartSessionHandler
    submitDetection *app.SubmitDetectionHandler
    confirmSession  *app.ConfirmSessionHandler
    cancelSession   *app.CancelSessionHandler
    
    // Test state
    currentSession  *domain.Session
    lastError       error
}

func TestTransactionFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeTransactionScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../../features/domain"},
            TestingT: t,
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned")
    }
}

func InitializeTransactionScenario(ctx *godog.ScenarioContext) {
    tc := &transactionTestContext{}
    
    ctx.Before(tc.beforeScenario)
    
    // Setup steps
    ctx.Step(`^a device exists with machine ID "([^"]*)"$`, tc.deviceExistsWithMachineID)
    ctx.Step(`^the following SKUs exist in catalog:$`, tc.followingSKUsExist)
    ctx.Step(`^an active session exists on device "([^"]*)"$`, tc.activeSessionExists)
    ctx.Step(`^a confirmed session exists on device "([^"]*)"$`, tc.confirmedSessionExists)
    ctx.Step(`^an active session with items exists on device "([^"]*)"$`, tc.activeSessionWithItemsExists)
    
    // Action steps
    ctx.Step(`^I start a session on device "([^"]*)"$`, tc.iStartSession)
    ctx.Step(`^I add the following detections:$`, tc.iAddDetections)
    ctx.Step(`^I try to add a detection for "([^"]*)"$`, tc.iTryToAddDetection)
    ctx.Step(`^I confirm the session$`, tc.iConfirmSession)
    ctx.Step(`^I cancel the session$`, tc.iCancelSession)
    
    // Assertion steps
    ctx.Step(`^the session should be in "([^"]*)" state$`, tc.sessionShouldBeInState)
    ctx.Step(`^the session should have (\d+) items$`, tc.sessionShouldHaveItems)
    ctx.Step(`^the total should be \$(\d+\.\d{2})$`, tc.totalShouldBe)
    ctx.Step(`^I should receive an error "([^"]*)"$`, tc.iShouldReceiveError)
    ctx.Step(`^the session should remain unchanged$`, tc.sessionShouldRemainUnchanged)
    ctx.Step(`^a "([^"]*)" event should be published$`, tc.eventShouldBePublished)
}

func (tc *transactionTestContext) beforeScenario(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
    // Initialize in-memory repositories
    tc.deviceRepo = NewInMemoryDeviceRepository()
    tc.skuRepo = NewInMemorySKURepository()
    tc.sessionRepo = NewInMemorySessionRepository()
    tc.eventBus = NewInMemoryEventBus()
    
    // Initialize application services
    tc.startSession = app.NewStartSessionHandler(
        tc.sessionRepo,
        tc.deviceRepo,
        tc.eventBus,
    )
    
    tc.submitDetection = app.NewSubmitDetectionHandler(
        tc.sessionRepo,
        tc.skuRepo,
        tc.eventBus,
    )
    
    tc.confirmSession = app.NewConfirmSessionHandler(
        tc.sessionRepo,
        tc.eventBus,
    )
    
    tc.cancelSession = app.NewCancelSessionHandler(
        tc.sessionRepo,
        tc.eventBus,
    )
    
    // Reset state
    tc.currentSession = nil
    tc.lastError = nil
    
    return ctx, nil
}

func (tc *transactionTestContext) deviceExistsWithMachineID(machineID string) error {
    device, err := devicedomain.NewDevice(machineID, "Test Device")
    if err != nil {
        return err
    }
    
    return tc.deviceRepo.Save(context.Background(), device)
}

func (tc *transactionTestContext) followingSKUsExist(table *godog.Table) error {
    for i, row := range table.Rows {
        if i == 0 {
            continue // Skip header
        }
        
        sku, err := catalogdomain.NewSKU(
            row.Cells[0].Value, // code
            row.Cells[1].Value, // name
            parseFloat(row.Cells[2].Value), // price
            parseInt(row.Cells[3].Value),   // weight
        )
        if err != nil {
            return err
        }
        
        if err := tc.skuRepo.Save(context.Background(), sku); err != nil {
            return err
        }
    }
    
    return nil
}

func (tc *transactionTestContext) iStartSession(machineID string) error {
    cmd := app.StartSessionCommand{
        MachineID: machineID,
    }
    
    result, err := tc.startSession.Handle(context.Background(), cmd)
    tc.lastError = err
    
    if err == nil {
        tc.currentSession, _ = tc.sessionRepo.FindByID(context.Background(), result.SessionID)
    }
    
    return nil
}

func (tc *transactionTestContext) sessionShouldBeInState(expectedState string) error {
    if tc.currentSession == nil {
        return fmt.Errorf("no current session")
    }
    
    actualState := tc.currentSession.State().String()
    if actualState != expectedState {
        return fmt.Errorf("expected state %s, got %s", expectedState, actualState)
    }
    
    return nil
}

func (tc *transactionTestContext) sessionShouldHaveItems(count int) error {
    if tc.currentSession == nil {
        return fmt.Errorf("no current session")
    }
    
    actualCount := len(tc.currentSession.Items())
    if actualCount != count {
        return fmt.Errorf("expected %d items, got %d", count, actualCount)
    }
    
    return nil
}

func (tc *transactionTestContext) totalShouldBe(amount float64) error {
    if tc.currentSession == nil {
        return fmt.Errorf("no current session")
    }
    
    actualTotal := tc.currentSession.Total().Amount()
    if actualTotal != amount {
        return fmt.Errorf("expected total $%.2f, got $%.2f", amount, actualTotal)
    }
    
    return nil
}

func (tc *transactionTestContext) eventShouldBePublished(eventType string) error {
    events := tc.eventBus.GetPublishedEvents()
    
    for _, event := range events {
        if event.Type() == eventType {
            return nil
        }
    }
    
    return fmt.Errorf("event %s was not published. Published: %v", eventType, events)
}
```

---

## Example 3: End-to-End Workflow

### Feature File

```gherkin
# features/e2e/checkout.feature
@e2e @checkout
Feature: Complete Checkout Flow
  As a customer
  I want to complete a purchase
  So that I can buy items

  Scenario: Successful checkout
    Given I am at a device with machine ID "DEVICE-001"
    And the following products are available:
      | code      | name       | price |
      | APPLE-001 | Fuji Apple | 2.50  |
      | APPLE-002 | Gala Apple | 2.30  |
    
    When I start a new shopping session
    And I place the following items on the scale:
      | product   | quantity |
      | APPLE-001 | 2        |
      | APPLE-002 | 1        |
    And I confirm my purchase
    
    Then I should see a total of $7.30
    And I should receive a receipt
    And the session should be marked as complete
    And an inventory update should be triggered
```

This example demonstrates a complete end-to-end test that exercises multiple bounded contexts.

---

## Helper Functions

```go
// test/support/helpers.go
package support

import (
    "strconv"
)

func parseInt(s string) int {
    v, _ := strconv.Atoi(s)
    return v
}

func parseFloat(s string) float64 {
    v, _ := strconv.ParseFloat(s, 64)
    return v
}

func parseBool(s string) bool {
    return s == "true" || s == "yes" || s == "1"
}
```

---

## In-Memory Test Doubles

```go
// test/support/in_memory_repository.go
package support

import (
    "context"
    "fmt"
    "sync"
)

type InMemorySessionRepository struct {
    mu       sync.RWMutex
    sessions map[string]*domain.Session
}

func NewInMemorySessionRepository() *InMemorySessionRepository {
    return &InMemorySessionRepository{
        sessions: make(map[string]*domain.Session),
    }
}

func (r *InMemorySessionRepository) Save(ctx context.Context, session *domain.Session) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.sessions[session.ID()] = session
    return nil
}

func (r *InMemorySessionRepository) FindByID(ctx context.Context, id string) (*domain.Session, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    session, exists := r.sessions[id]
    if !exists {
        return nil, fmt.Errorf("session not found")
    }
    
    return session, nil
}

func (r *InMemorySessionRepository) Clear() {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    r.sessions = make(map[string]*domain.Session)
}
```

```go
// test/support/in_memory_event_bus.go
package support

import (
    "sync"
    
    "myservice/internal/shared/events"
)

type InMemoryEventBus struct {
    mu     sync.RWMutex
    events []events.DomainEvent
}

func NewInMemoryEventBus() *InMemoryEventBus {
    return &InMemoryEventBus{
        events: make([]events.DomainEvent, 0),
    }
}

func (b *InMemoryEventBus) Publish(ctx context.Context, events ...events.DomainEvent) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    b.events = append(b.events, events...)
    return nil
}

func (b *InMemoryEventBus) GetPublishedEvents() []events.DomainEvent {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    return append([]events.DomainEvent{}, b.events...)
}

func (b *InMemoryEventBus) Clear() {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    b.events = make([]events.DomainEvent, 0)
}
```

These examples provide a solid foundation for implementing BDD tests with Godog in your Go projects.
