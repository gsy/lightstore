# BDD Tests with Godog

This directory contains Behavior-Driven Development (BDD) tests for the vending machine backend server using Godog (Cucumber for Go).

## Structure

```
test/
├── bdd_test.go              # Main test suite entry point
├── common_steps.go          # Common step definitions (API, assertions)
├── catalog_steps.go         # Catalog context step definitions
├── device_steps.go          # Device context step definitions
├── transaction_steps.go     # Transaction context step definitions
├── support/                 # Test support utilities
│   ├── test_context.go      # Shared test context
│   └── test_server.go       # Test server setup
└── README.md                # This file

features/
├── catalog/
│   └── manage_skus.feature  # SKU management scenarios
├── device/
│   └── device_management.feature  # Device registration scenarios
└── transaction/
    └── shopping_session.feature   # Shopping session scenarios
```

## Prerequisites

1. **PostgreSQL database** for testing:
   ```bash
   # Create test database
   createdb vending_test
   
   # Or use Docker
   docker run -d \
     --name postgres-test \
     -e POSTGRES_USER=vending \
     -e POSTGRES_PASSWORD=vending \
     -e POSTGRES_DB=vending_test \
     -p 5432:5432 \
     postgres:15
   ```

2. **Install dependencies**:
   ```bash
   cd server
   go mod download
   ```

## Running Tests

### Run all BDD tests

```bash
cd server
go test -v ./test/...
```

### Run with specific tags

```bash
# Run only smoke tests
go test -v ./test/... -godog.tags="@smoke"

# Run catalog tests only
go test -v ./test/... -godog.tags="@catalog"

# Run all except error-handling tests
go test -v ./test/... -godog.tags="~@error-handling"

# Combine tags
go test -v ./test/... -godog.tags="@api && @smoke"
```

### Run specific feature file

```bash
go test -v ./test/... -godog.paths="../features/catalog/manage_skus.feature"
```

### Different output formats

```bash
# Pretty format (default)
go test -v ./test/... -godog.format=pretty

# Progress dots
go test -v ./test/... -godog.format=progress

# JUnit XML (for CI)
go test -v ./test/... -godog.format=junit > report.xml

# Cucumber JSON (for reporting tools)
go test -v ./test/... -godog.format=cucumber > report.json
```

## Environment Variables

- `DATABASE_URL`: PostgreSQL connection string (default: `postgres://vending:vending@localhost:5432/vending_test?sslmode=disable`)

Example:
```bash
DATABASE_URL="postgres://user:pass@localhost:5432/testdb?sslmode=disable" go test -v ./test/...
```

## Writing New Tests

### 1. Create a Feature File

Create a new `.feature` file in the appropriate directory under `features/`:

```gherkin
@api @myfeature
Feature: My New Feature
  As a user
  I want to do something
  So that I can achieve a goal

  Background:
    Given the API server is running
    And the database is clean

  Scenario: Do something successfully
    When I perform an action
    Then the result should be successful
```

### 2. Implement Step Definitions

Add step definitions to the appropriate `*_steps.go` file or create a new one:

```go
func iPerformAnAction() error {
    // Implementation
    return testContext.SendRequest("POST", "/api/v1/action", nil)
}

func theResultShouldBeSuccessful() error {
    if testContext.LastResponse.StatusCode != 200 {
        return fmt.Errorf("expected 200, got %d", testContext.LastResponse.StatusCode)
    }
    return nil
}
```

### 3. Register Steps

Add step registration in `bdd_test.go`:

```go
ctx.Step(`^I perform an action$`, iPerformAnAction)
ctx.Step(`^the result should be successful$`, theResultShouldBeSuccessful)
```

## Test Patterns

### Testing API Endpoints

```gherkin
When I send a POST request to "/api/v1/resource" with body:
  """
  {
    "field": "value"
  }
  """
Then the response status should be 201
And the response should contain field "id"
```

### Testing with Data Tables

```gherkin
Given the following resources exist:
  | name    | value |
  | Item 1  | 100   |
  | Item 2  | 200   |
```

### Testing Error Cases

```gherkin
@error-handling
Scenario: Handle invalid input
  When I send invalid data
  Then the response status should be 400
  And the response should contain error "validation failed"
```

### Using Scenario Outlines

```gherkin
Scenario Outline: Validate fields
  When I create a resource with <field> set to <value>
  Then the response status should be <status>

  Examples:
    | field | value | status |
    | name  | ""    | 400    |
    | value | -1    | 422    |
```

## CI/CD Integration

### GitHub Actions Example

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
          POSTGRES_USER: vending
          POSTGRES_PASSWORD: vending
          POSTGRES_DB: vending_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
    
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run BDD tests
        working-directory: ./server
        run: go test -v ./test/...
        env:
          DATABASE_URL: postgres://vending:vending@localhost:5432/vending_test?sslmode=disable
```

## Troubleshooting

### Database connection issues

```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Test connection
psql -h localhost -U vending -d vending_test
```

### Tests fail with "table does not exist"

The migrations should run automatically. If they don't:

```bash
# Manually run migrations
cd server
go run cmd/server/main.go
```

### Port already in use

The test server uses a random port via `httptest.NewServer()`, so port conflicts shouldn't occur.

## Best Practices

1. **Keep scenarios focused**: Each scenario should test one behavior
2. **Use Background wisely**: Put common setup in Background
3. **Write from user perspective**: Use business language, not technical details
4. **Tag appropriately**: Use tags like `@smoke`, `@wip`, `@slow` for test organization
5. **Clean state**: Always start with a clean database state
6. **Avoid UI coupling**: Test business logic, not implementation details
7. **Make steps reusable**: Write generic steps that work across scenarios

## Resources

- [Godog Documentation](https://github.com/cucumber/godog)
- [Gherkin Reference](https://cucumber.io/docs/gherkin/reference/)
- [BDD Best Practices](https://cucumber.io/docs/bdd/)
