# BDD Testing with Godog - Complete Guide

This document provides a comprehensive overview of the BDD testing setup for the vending machine backend server.

## 📋 Table of Contents

- [Overview](#overview)
- [What We're Testing](#what-were-testing)
- [Quick Start](#quick-start)
- [Test Coverage](#test-coverage)
- [Architecture](#architecture)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
- [CI/CD Integration](#cicd-integration)
- [Best Practices](#best-practices)

## Overview

We use **Godog** (Cucumber for Go) to write executable specifications in plain language (Gherkin) that serve as both documentation and automated tests. This approach bridges the gap between business requirements and technical implementation.

### Why BDD?

- **Living Documentation**: Tests are written in business language that anyone can understand
- **Collaboration**: Product owners, developers, and QA can all contribute to test scenarios
- **Confidence**: High-level integration tests ensure the system works end-to-end
- **Regression Prevention**: Automated tests catch breaking changes early

## What We're Testing

Our BDD tests cover three bounded contexts:

### 1. Catalog Context (`@catalog`)
- ✅ Create SKUs with validation
- ✅ List all SKUs
- ✅ List active SKUs only
- ✅ Get SKU by ID
- ✅ Handle duplicate SKU codes
- ✅ Validate required fields

**Feature File**: `features/catalog/manage_skus.feature`

### 2. Device Context (`@device`)
- ✅ Register new devices
- ✅ Handle duplicate device registration
- ✅ Retrieve active SKUs for ML model sync
- ✅ Validate device fields

**Feature File**: `features/device/device_management.feature`

### 3. Transaction Context (`@transaction`)
- ✅ Start shopping sessions
- ✅ Submit item detections
- ✅ Get session details
- ✅ Confirm sessions (complete purchase)
- ✅ Cancel sessions
- ✅ Handle non-existent devices/sessions
- ✅ Prevent operations on completed sessions
- ✅ Validate session state transitions

**Feature File**: `features/transaction/shopping_session.feature`

## Quick Start

See [QUICKSTART.md](QUICKSTART.md) for a 5-minute getting started guide.

```bash
# 1. Start test database
docker run -d --name postgres-test \
  -e POSTGRES_USER=vending \
  -e POSTGRES_PASSWORD=vending \
  -e POSTGRES_DB=vending_test \
  -p 5432:5432 postgres:15

# 2. Run tests
make test-bdd

# 3. Run smoke tests only
make test-bdd-smoke
```

## Test Coverage

### Current Coverage

| Context     | Scenarios | Steps | Coverage |
|-------------|-----------|-------|----------|
| Catalog     | 6         | 25+   | Core CRUD operations |
| Device      | 4         | 15+   | Registration & sync |
| Transaction | 9         | 40+   | Full session lifecycle |
| **Total**   | **19**    | **80+** | **High-value paths** |

### Test Tags

- `@smoke` - Critical path tests (run on every commit)
- `@api` - API integration tests
- `@catalog` - Catalog context tests
- `@device` - Device context tests
- `@transaction` - Transaction context tests
- `@error-handling` - Error case tests
- `@validation` - Input validation tests

## Architecture

### Test Structure

```
server/
├── features/                    # Gherkin feature files
│   ├── catalog/
│   ├── device/
│   └── transaction/
├── test/
│   ├── bdd_test.go             # Main test suite
│   ├── common_steps.go         # Shared step definitions
│   ├── catalog_steps.go        # Catalog-specific steps
│   ├── device_steps.go         # Device-specific steps
│   ├── transaction_steps.go    # Transaction-specific steps
│   └── support/
│       ├── test_context.go     # Shared test state
│       └── test_server.go      # Test server setup
```

### Test Flow

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Feature File (Gherkin)                                    │
│    Given the API server is running                           │
│    When I create a SKU with code "APPLE-001"                 │
│    Then the response status should be 201                    │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. Step Definitions (Go)                                     │
│    func iCreateSKU(code string) error {                      │
│        return testContext.SendRequest(...)                   │
│    }                                                          │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Test Context (State Management)                           │
│    - HTTP client                                              │
│    - Test server                                              │
│    - Database connection                                      │
│    - Request/response storage                                 │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Application (Real Server)                                 │
│    - All bounded contexts                                     │
│    - Real database (test DB)                                  │
│    - Full dependency injection                                │
└─────────────────────────────────────────────────────────────┘
```

### Test Isolation

Each scenario:
1. Starts with a clean database
2. Gets a fresh test server instance
3. Maintains isolated state in `TestContext`
4. Cleans up after completion

## Running Tests

### Basic Commands

```bash
# All tests
make test-bdd

# Specific context
make test-bdd-catalog
make test-bdd-device
make test-bdd-transaction

# Smoke tests only
make test-bdd-smoke

# Different output formats
make test-bdd-progress    # Progress dots
make test-bdd-junit       # JUnit XML
make test-bdd-cucumber    # Cucumber JSON
```

### Advanced Usage

```bash
# Run specific feature
cd server
go test -v ./test/... -godog.paths="../features/catalog/manage_skus.feature"

# Run specific scenario (by line number)
go test -v ./test/... -godog.paths="../features/catalog/manage_skus.feature:9"

# Combine tags
go test -v ./test/... -godog.tags="@api && @smoke"
go test -v ./test/... -godog.tags="@catalog || @device"
go test -v ./test/... -godog.tags="~@error-handling"  # Exclude

# Custom database
DATABASE_URL="postgres://user:pass@host:5432/db" make test-bdd

# Verbose output
go test -v ./test/... -godog.format=pretty -godog.no-colors=false
```

## Writing Tests

### 1. Write the Feature

```gherkin
@api @catalog
Feature: Deactivate SKU
  As a catalog manager
  I want to deactivate SKUs
  So that they are no longer available

  Background:
    Given the API server is running
    And the database is clean

  Scenario: Deactivate active SKU
    Given a SKU exists with code "APPLE-001"
    When I deactivate the SKU "APPLE-001"
    Then the response status should be 200
    And the SKU should be inactive
```

### 2. Run to See Undefined Steps

```bash
make test-bdd-catalog
```

Output will show:
```
Step is undefined: I deactivate the SKU "APPLE-001"

You can implement step definitions for undefined steps with these snippets:

func iDeactivateTheSKU(arg1 string) error {
    return godog.ErrPending
}
```

### 3. Implement Step Definitions

Add to `catalog_steps.go`:

```go
func iDeactivateTheSKU(code string) error {
    skuID := testContext.CreatedSKUs[code]
    return testContext.SendRequest("POST", 
        fmt.Sprintf("/api/v1/skus/%s/deactivate", skuID), nil)
}

func theSKUShouldBeInactive() error {
    response, err := testContext.GetResponseJSON()
    if err != nil {
        return err
    }
    
    if active, ok := response["active"].(bool); !ok || active {
        return fmt.Errorf("expected SKU to be inactive")
    }
    
    return nil
}
```

### 4. Register Steps

Add to `bdd_test.go`:

```go
ctx.Step(`^I deactivate the SKU "([^"]*)"$`, iDeactivateTheSKU)
ctx.Step(`^the SKU should be inactive$`, theSKUShouldBeInactive)
```

### 5. Implement the Feature

Add the deactivate endpoint to your application.

### 6. Run Tests Again

```bash
make test-bdd-catalog
```

✅ Tests should pass!

## CI/CD Integration

### GitHub Actions

We have two workflows:

1. **Full BDD Test Suite** (`.github/workflows/bdd-tests.yml`)
   - Runs on push/PR to main/develop
   - Executes all BDD tests
   - Generates test reports
   - Publishes results

2. **Smoke Tests** (same file, separate job)
   - Quick validation
   - Runs critical path tests only
   - Fast feedback loop

### Running in CI

```yaml
- name: Run BDD tests
  run: make test-bdd
  env:
    DATABASE_URL: postgres://vending:vending@localhost:5432/vending_test?sslmode=disable
```

### Test Reports

Reports are generated in multiple formats:
- **JUnit XML**: For CI integration
- **Cucumber JSON**: For reporting tools
- **Pretty**: For human reading

## Best Practices

### ✅ DO

1. **Write from user perspective**
   ```gherkin
   # Good
   When I create a SKU with code "APPLE-001"
   
   # Bad
   When I POST to /api/v1/skus with JSON payload
   ```

2. **Keep scenarios focused**
   - One scenario = one behavior
   - 5-10 steps maximum
   - Use Background for common setup

3. **Use descriptive names**
   ```gherkin
   # Good
   Scenario: Prevent duplicate SKU codes
   
   # Bad
   Scenario: Test SKU creation
   ```

4. **Tag appropriately**
   ```gherkin
   @smoke @catalog
   Scenario: Create SKU successfully
   ```

5. **Make steps reusable**
   ```gherkin
   # Reusable across scenarios
   Given a SKU exists with code "APPLE-001"
   ```

### ❌ DON'T

1. **Don't couple to implementation**
   ```gherkin
   # Bad - too technical
   When I call the CreateSKUHandler with parameters
   
   # Good - business language
   When I create a SKU with the following details
   ```

2. **Don't make scenarios too long**
   - Split into multiple scenarios
   - Use Background for setup

3. **Don't test every edge case in BDD**
   - Use unit tests for edge cases
   - BDD for high-value integration paths

4. **Don't share state between scenarios**
   - Each scenario should be independent
   - Use Background for setup, not previous scenarios

5. **Don't skip the Background**
   ```gherkin
   # Good
   Background:
     Given the API server is running
     And the database is clean
   
   # Bad - repeating in every scenario
   Scenario: Test 1
     Given the API server is running
     And the database is clean
     ...
   ```

## Troubleshooting

### Common Issues

1. **Database connection failed**
   ```bash
   # Check PostgreSQL is running
   docker ps | grep postgres-test
   
   # Restart if needed
   docker start postgres-test
   ```

2. **Port already in use**
   - Test server uses random ports (httptest)
   - Should not conflict

3. **Tests are slow**
   ```bash
   # Run smoke tests only
   make test-bdd-smoke
   
   # Or specific context
   make test-bdd-catalog
   ```

4. **Step not matching**
   - Check regex patterns
   - Escape special characters
   - Use exact quotes from feature file

## Resources

- [Godog Documentation](https://github.com/cucumber/godog)
- [Gherkin Reference](https://cucumber.io/docs/gherkin/reference/)
- [BDD Best Practices](https://cucumber.io/docs/bdd/)
- [BDD-Godog Skill](.claude/skills/bdd-godog/SKILL.md)
- [Practical Examples](.claude/skills/bdd-godog/references/practical-examples.md)

## Support

For questions or issues:
1. Check [QUICKSTART.md](QUICKSTART.md)
2. Review [README.md](README.md)
3. Check existing feature files for examples
4. Review step definitions for patterns

---

**Happy Testing!** 🎉

Remember: Good BDD tests are living documentation that helps everyone understand what the system does and gives confidence to make changes.
