---
name: bdd-godog
description: >
  Implement Behavior-Driven Development (BDD) testing in Go using Godog, the official Cucumber
  framework for Golang. Use this skill when the user asks to write BDD tests, create feature files,
  implement step definitions, write Gherkin scenarios, or set up acceptance testing. Also trigger
  when the user mentions "cucumber", "gherkin", "feature files", "given-when-then", "acceptance tests",
  "behavior tests", or asks to test business requirements in plain language. This skill helps bridge
  the gap between business requirements and technical implementation through executable specifications.
---

# Golang BDD Testing with Godog

This skill enables you to write executable specifications using Gherkin syntax and implement them
in Go using Godog. The goal is to create living documentation that serves as both specification
and automated tests, facilitating collaboration between technical and non-technical stakeholders.

---

## Mental Model

BDD follows a three-layer approach:

```
┌─────────────────────────────────────────────┐
│  Feature Files (.feature)                    │
│  Human-readable specifications in Gherkin    │
│  Given-When-Then scenarios                   │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│  Step Definitions (*_test.go)                │
│  Go code that implements each step           │
│  Maps Gherkin phrases to test code           │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│  Application Code                            │
│  The actual system being tested              │
│  Domain logic, APIs, services                │
└─────────────────────────────────────────────┘
```

**The BDD cycle**: Write feature → Run (see undefined steps) → Implement steps → Run (see failures) → Implement code → Run (see success)

---

## Installation

```bash
go get github.com/cucumber/godog/cmd/godog
```

Or add to your `go.mod`:

```go
require github.com/cucumber/godog v0.14.0
```

---

## Project Structure

Organize BDD tests alongside your application code:

```
my-service/
├── features/                         # Feature files directory
│   ├── catalog/
│   │   ├── create_sku.feature        # Feature: Create SKU
│   │   └── list_skus.feature         # Feature: List SKUs
│   ├── device/
│   │   └── register_device.feature
│   └── transaction/
│       ├── start_session.feature
│       └── submit_detection.feature
│
├── internal/
│   ├── catalog/
│   │   ├── domain/
│   │   ├── app/
│   │   ├── infra/
│   │   └── bdd_test.go               # Step definitions for catalog
│   ├── device/
│   │   └── bdd_test.go
│   └── transaction/
│       └── bdd_test.go
│
├── test/
│   ├── support/                      # Shared test utilities
│   │   ├── api_client.go             # HTTP client helpers
│   │   ├── database.go               # DB setup/teardown
│   │   └── context.go                # Shared test context
│   └── bdd_suite_test.go             # Main test suite entry point
│
└── go.mod
```

**Alternative: Centralized approach**

```
my-service/
├── features/                         # All feature files
│   ├── catalog.feature
│   ├── device.feature
│   └── transaction.feature
│
├── test/
│   ├── bdd/
│   │   ├── steps/                    # All step definitions
│   │   │   ├── catalog_steps.go
│   │   │   ├── device_steps.go
│   │   │   └── transaction_steps.go
│   │   ├── support/                  # Test helpers
│   │   └── suite_test.go             # Suite configuration
│   └── integration/                  # Traditional integration tests
│
└── internal/                         # Application code
```

---

## Gherkin Syntax — Quick Reference

### Feature File Structure

```gherkin
# features/catalog/create_sku.feature
Feature: Create SKU
  As a catalog manager
  I want to create new SKUs
  So that I can manage product inventory

  Background:
    Given the catalog service is running
    And the database is clean

  Scenario: Create a valid SKU
    Given I am authenticated as a catalog manager
    When I create a SKU with the following details:
      | code        | name          | price | weight |
      | APPLE-001   | Fuji Apple    | 2.50  | 150    |
    Then the SKU should be created successfully
    And the SKU should be retrievable by code "APPLE-001"

  Scenario: Reject SKU with duplicate code
    Given a SKU exists with code "APPLE-001"
    When I create a SKU with code "APPLE-001"
    Then I should receive a "duplicate code" error
    And no new SKU should be created

  Scenario Outline: Validate SKU fields
    When I create a SKU with <field> set to <value>
    Then I should receive a "<error>" error

    Examples:
      | field  | value | error                |
      | code   | ""    | code required        |
      | name   | ""    | name required        |
      | price  | -1    | price must be positive |
      | weight | 0     | weight must be positive |
```

### Gherkin Keywords

| Keyword | Purpose | Example |
|---|---|---|
| `Feature:` | High-level description | `Feature: User Authentication` |
| `Background:` | Steps run before each scenario | `Background: Given the database is clean` |
| `Scenario:` | Single test case | `Scenario: Login with valid credentials` |
| `Scenario Outline:` | Parameterized scenario | `Scenario Outline: Validate <field>` |
| `Examples:` | Data table for outline | `Examples: \| field \| value \|` |
| `Given` | Precondition/setup | `Given a user exists` |
| `When` | Action/trigger | `When I submit the login form` |
| `Then` | Expected outcome | `Then I should see the dashboard` |
| `And` / `But` | Additional steps | `And I should receive a welcome email` |

---

## Step Definitions — Implementation Patterns

### Basic Step Definition

```go
// internal/catalog/bdd_test.go
package catalog_test

import (
    "context"
    "fmt"
    "testing"

    "github.com/cucumber/godog"
)

// Test context holds state between steps
type catalogTestContext struct {
    apiClient    *APIClient
    lastResponse *http.Response
    lastError    error
    createdSKU   *SKU
}

func TestFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeScenario,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../../features/catalog"},
            TestingT: t,
        },
    }

    if suite.Run() != 0 {
        t.Fatal("non-zero status returned, failed to run feature tests")
    }
}

func InitializeScenario(ctx *godog.ScenarioContext) {
    tc := &catalogTestContext{
        apiClient: NewAPIClient(),
    }

    // Register step definitions
    ctx.Step(`^I am authenticated as a catalog manager$`, tc.iAmAuthenticatedAs)
    ctx.Step(`^I create a SKU with the following details:$`, tc.iCreateSKUWithDetails)
    ctx.Step(`^the SKU should be created successfully$`, tc.skuShouldBeCreatedSuccessfully)
    ctx.Step(`^the SKU should be retrievable by code "([^"]*)"$`, tc.skuShouldBeRetrievableByCode)

    // Hooks
    ctx.Before(tc.beforeScenario)
    ctx.After(tc.afterScenario)
}

func (tc *catalogTestContext) beforeScenario(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
    // Clean database, reset state
    tc.apiClient.CleanDatabase()
    return ctx, nil
}

func (tc *catalogTestContext) afterScenario(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
    // Cleanup resources
    return ctx, nil
}
```

### Step Implementation Examples

```go
// Simple assertion step
func (tc *catalogTestContext) iAmAuthenticatedAs(role string) error {
    token, err := tc.apiClient.Authenticate(role)
    if err != nil {
        return fmt.Errorf("authentication failed: %w", err)
    }
    tc.apiClient.SetAuthToken(token)
    return nil
}

// Step with data table
func (tc *catalogTestContext) iCreateSKUWithDetails(table *godog.Table) error {
    // Parse table into struct
    sku := parseSKUFromTable(table)
    
    resp, err := tc.apiClient.CreateSKU(sku)
    tc.lastResponse = resp
    tc.lastError = err
    
    if err == nil && resp.StatusCode == 201 {
        tc.createdSKU = resp.Body.SKU
    }
    
    return nil  // Don't return error here; validate in Then step
}

// Step with regex capture
func (tc *catalogTestContext) skuShouldBeRetrievableByCode(code string) error {
    sku, err := tc.apiClient.GetSKUByCode(code)
    if err != nil {
        return fmt.Errorf("failed to retrieve SKU: %w", err)
    }
    
    if sku.Code != code {
        return fmt.Errorf("expected SKU code %s, got %s", code, sku.Code)
    }
    
    return nil
}

// Assertion step
func (tc *catalogTestContext) skuShouldBeCreatedSuccessfully() error {
    if tc.lastError != nil {
        return fmt.Errorf("expected success but got error: %w", tc.lastError)
    }
    
    if tc.lastResponse.StatusCode != 201 {
        return fmt.Errorf("expected status 201, got %d", tc.lastResponse.StatusCode)
    }
    
    if tc.createdSKU == nil {
        return fmt.Errorf("no SKU was created")
    }
    
    return nil
}

// Error validation step
func (tc *catalogTestContext) iShouldReceiveError(expectedError string) error {
    if tc.lastError == nil {
        return fmt.Errorf("expected error %q but got none", expectedError)
    }
    
    if !strings.Contains(tc.lastError.Error(), expectedError) {
        return fmt.Errorf("expected error containing %q, got %q", expectedError, tc.lastError.Error())
    }
    
    return nil
}
```

### Regex Patterns for Step Matching

```go
// Exact string match
ctx.Step(`^I am on the home page$`, tc.iAmOnHomePage)

// Capture string in quotes
ctx.Step(`^I enter "([^"]*)" in the search box$`, tc.iEnterInSearchBox)

// Capture number
ctx.Step(`^there are (\d+) items in the cart$`, tc.thereAreItemsInCart)

// Capture decimal
ctx.Step(`^the total price is \$(\d+\.\d{2})$`, tc.totalPriceIs)

// Optional word
ctx.Step(`^the item should( not)? be visible$`, tc.itemShouldBeVisible)

// Multiple captures
ctx.Step(`^I transfer \$(\d+\.\d{2}) from "([^"]*)" to "([^"]*)"$`, tc.iTransferMoney)
```

---

## Running Tests

### Command Line

```bash
# Run all features
godog

# Run specific feature
godog features/catalog/create_sku.feature

# Run with tags
godog --tags=@wip
godog --tags='@catalog && @smoke'
godog --tags='~@skip'  # Exclude @skip

# Different formats
godog --format=pretty
godog --format=progress
godog --format=junit > report.xml
godog --format=cucumber > report.json

# Run specific scenario
godog features/catalog/create_sku.feature:12  # Line number
```

### Using Go Test

```bash
# Run as regular Go tests
go test ./test/bdd/...

# With verbose output
go test -v ./test/bdd/...

# Run specific test
go test -v ./test/bdd -run TestFeatures
```

### Tags for Test Organization

```gherkin
@smoke @catalog
Feature: Create SKU

  @happy-path
  Scenario: Create valid SKU
    # ...

  @error-handling @wip
  Scenario: Handle duplicate code
    # ...
```

```go
// Run only @smoke tests
Options: &godog.Options{
    Tags: "@smoke",
}
```

---

## Best Practices

### Writing Good Gherkin

**DO:**
- Write from user perspective, not implementation
- Use business language, not technical jargon
- Keep scenarios focused on one behavior
- Use Background for common setup
- Use Scenario Outline for similar cases with different data

**DON'T:**
- Couple scenarios to UI elements (`When I click button #submit-btn`)
- Write implementation details (`When I call POST /api/v1/skus`)
- Make scenarios too long (>10 steps is a smell)
- Repeat setup in every scenario (use Background)

**Good:**
```gherkin
Scenario: Customer purchases item
  Given I have an item in my cart
  When I complete the checkout process
  Then I should receive an order confirmation
```

**Bad:**
```gherkin
Scenario: Customer purchases item
  Given I navigate to "http://localhost:8080/products"
  And I click the "Add to Cart" button with id "add-cart-123"
  And I click the cart icon in the top right
  When I fill in the form field "email" with "test@example.com"
  And I click the "Submit Order" button
  Then the database should contain a row in orders table
```

### Step Definition Guidelines

**Keep steps reusable:**
```go
// Good - reusable across scenarios
ctx.Step(`^a SKU exists with code "([^"]*)"$`, tc.skuExistsWithCode)

// Bad - too specific
ctx.Step(`^a SKU exists with code "APPLE-001" and name "Fuji Apple"$`, tc.specificSKU)
```

**Separate actions from assertions:**
```go
// Action step - don't assert
func (tc *catalogTestContext) iCreateSKU(code string) error {
    _, err := tc.apiClient.CreateSKU(code)
    tc.lastError = err
    return nil  // Store error, don't return it
}

// Assertion step - validate outcome
func (tc *catalogTestContext) skuShouldBeCreated() error {
    if tc.lastError != nil {
        return fmt.Errorf("expected success: %w", tc.lastError)
    }
    return nil
}
```

**Use test context for state:**
```go
type testContext struct {
    // API clients
    apiClient *APIClient
    
    // State from previous steps
    currentUser   *User
    createdSKU    *SKU
    lastResponse  *http.Response
    lastError     error
    
    // Test data
    testData map[string]interface{}
}
```

### Integration with Hexagonal Architecture

When testing a DDD/Hexagonal application:

```go
type testContext struct {
    // Real infrastructure for integration tests
    db          *sql.DB
    httpServer  *httptest.Server
    
    // Or use in-memory adapters for faster tests
    repo        *InMemoryRepository
    eventBus    *InMemoryEventBus
    
    // Application services
    createSKU   *app.CreateSKUHandler
    querySKU    *app.SKUQueryService
}

func (tc *testContext) beforeScenario(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
    // Option 1: Test through HTTP (full integration)
    tc.httpServer = startTestServer()
    
    // Option 2: Test application layer directly (faster)
    tc.repo = NewInMemoryRepository()
    tc.createSKU = app.NewCreateSKUHandler(tc.repo, tc.eventBus)
    
    return ctx, nil
}
```

### Hooks and Lifecycle

```go
func InitializeScenario(ctx *godog.ScenarioContext) {
    tc := &testContext{}
    
    // Before hooks
    ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
        // Runs before each scenario
        tc.cleanDatabase()
        return ctx, nil
    })
    
    ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
        // Runs after each scenario
        if err != nil {
            tc.captureDebugInfo()
        }
        tc.cleanup()
        return ctx, nil
    })
    
    // Step-level hooks
    ctx.BeforeStep(func(ctx context.Context, st *godog.Step) (context.Context, error) {
        // Runs before each step
        return ctx, nil
    })
    
    ctx.AfterStep(func(ctx context.Context, st *godog.Step, err error) (context.Context, error) {
        // Runs after each step
        if err != nil {
            tc.logStepFailure(st, err)
        }
        return ctx, nil
    })
    
    // Register steps
    tc.registerSteps(ctx)
}
```

---

## Common Patterns

### API Testing

```go
type apiTestContext struct {
    baseURL      string
    client       *http.Client
    authToken    string
    lastResponse *http.Response
    lastBody     []byte
}

func (tc *apiTestContext) iSendRequestTo(method, path string) error {
    req, _ := http.NewRequest(method, tc.baseURL+path, nil)
    if tc.authToken != "" {
        req.Header.Set("Authorization", "Bearer "+tc.authToken)
    }
    
    resp, err := tc.client.Do(req)
    tc.lastResponse = resp
    if resp != nil {
        tc.lastBody, _ = io.ReadAll(resp.Body)
        resp.Body.Close()
    }
    
    return err
}

func (tc *apiTestContext) theResponseCodeShouldBe(code int) error {
    if tc.lastResponse.StatusCode != code {
        return fmt.Errorf("expected %d, got %d: %s", 
            code, tc.lastResponse.StatusCode, string(tc.lastBody))
    }
    return nil
}
```

### Database State Verification

```go
func (tc *testContext) skuShouldExistInDatabase(code string) error {
    var count int
    err := tc.db.QueryRow("SELECT COUNT(*) FROM skus WHERE code = $1", code).Scan(&count)
    if err != nil {
        return err
    }
    
    if count == 0 {
        return fmt.Errorf("SKU with code %s not found in database", code)
    }
    
    return nil
}
```

### Event Verification

```go
type eventTestContext struct {
    eventBus *InMemoryEventBus
}

func (tc *eventTestContext) eventShouldBePublished(eventType string) error {
    events := tc.eventBus.GetPublishedEvents()
    
    for _, event := range events {
        if event.Type() == eventType {
            return nil
        }
    }
    
    return fmt.Errorf("event %s was not published", eventType)
}
```

---

## Testing Strategy

### Test Pyramid with BDD

```
        ┌─────────────┐
        │  BDD Tests  │  ← Few, high-value scenarios
        │  (Godog)    │     Full system integration
        └─────────────┘
            ↑
    ┌───────────────────┐
    │ Integration Tests │  ← More tests, focused on
    │  (Go test)        │     component integration
    └───────────────────┘
            ↑
┌─────────────────────────────┐
│     Unit Tests              │  ← Most tests, fast
│     (Go test)               │     Pure domain logic
└─────────────────────────────┘
```

**When to use BDD tests:**
- Critical business flows (checkout, payment, registration)
- Cross-cutting scenarios (authentication, authorization)
- Acceptance criteria from stakeholders
- Regression tests for high-impact bugs

**When NOT to use BDD tests:**
- Low-level implementation details
- Edge cases better covered by unit tests
- Performance testing
- Every possible scenario (too slow)

---

## CI/CD Integration

### GitHub Actions

```yaml
name: BDD Tests

on: [push, pull_request]

jobs:
  bdd:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run BDD tests
        run: go test -v ./test/bdd/...
        env:
          DATABASE_URL: postgres://postgres:test@localhost:5432/test?sslmode=disable
      
      - name: Generate report
        if: always()
        run: |
          godog --format=cucumber:report.json
          godog --format=junit:report.xml
      
      - name: Upload report
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: bdd-report
          path: report.*
```

---

## Troubleshooting

### Undefined Steps

When you see:
```
Step is undefined: I create a SKU with code "APPLE-001"
```

Godog provides the snippet:
```go
ctx.Step(`^I create a SKU with code "([^"]*)"$`, iCreateASKUWithCode)
```

Copy and implement it.

### Step Not Matching

If your step isn't matching, check:
- Regex special characters need escaping: `\$`, `\(`, `\)`
- Quotes in Gherkin must match regex: `"([^"]*)"`
- Numbers: `(\d+)` for integers, `(\d+\.\d+)` for decimals

### Flaky Tests

Common causes:
- Async operations not properly awaited
- Shared state between scenarios (use `Before` hook to clean)
- Time-dependent logic (use fixed time in tests)
- External dependencies (use test doubles or containers)

---

## Reference Files

For complete examples, see:
- [Godog Official Docs](https://github.com/cucumber/godog)
- [Cucumber Gherkin Reference](https://cucumber.io/docs/gherkin/reference/)

## External Resources

- [Godog GitHub Repository](https://github.com/cucumber/godog) — Official implementation
- [Cucumber Documentation](https://cucumber.io/docs/) — Gherkin syntax and BDD practices
- [BDD in Action](https://www.manning.com/books/bdd-in-action) — Comprehensive BDD guide
