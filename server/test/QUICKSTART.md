# BDD Testing Quick Start Guide

Get started with BDD testing for the vending machine backend in 5 minutes.

## 1. Setup Test Database

```bash
# Option A: Using Docker (recommended)
docker run -d \
  --name postgres-test \
  -e POSTGRES_USER=vending \
  -e POSTGRES_PASSWORD=vending \
  -e POSTGRES_DB=vending_test \
  -p 5432:5432 \
  postgres:15

# Option B: Using local PostgreSQL
createdb vending_test
```

## 2. Install Dependencies

```bash
cd server
go mod download
```

## 3. Run Your First BDD Test

```bash
# Run all BDD tests
make test-bdd

# Or directly with go test
cd server
go test -v ./test/...
```

You should see output like:

```
Feature: Manage SKUs
  As a catalog manager
  I want to manage product SKUs
  So that I can maintain the product inventory

  Scenario: Create a new SKU                    # features/catalog/manage_skus.feature:9
    Given the API server is running             # common_steps.go:10
    And the database is clean                   # common_steps.go:15
    When I create a SKU with the following details: # catalog_steps.go:8
      | code      | name       | price_cents | weight_grams |
      | APPLE-001 | Fuji Apple | 250         | 150          |
    Then the response status should be 201      # common_steps.go:25
    And the response should contain field "id"  # common_steps.go:35

3 scenarios (3 passed)
15 steps (15 passed)
```

## 4. Run Specific Tests

```bash
# Run only smoke tests (quick validation)
make test-bdd-smoke

# Run catalog tests only
make test-bdd-catalog

# Run device tests only
make test-bdd-device

# Run transaction tests only
make test-bdd-transaction
```

## 5. Explore the Features

Check out the feature files to see what's being tested:

```bash
# Catalog features
cat features/catalog/manage_skus.feature

# Device features
cat features/device/device_management.feature

# Transaction features
cat features/transaction/shopping_session.feature
```

## 6. Write Your First Test

### Step 1: Create a feature file

Create `features/catalog/deactivate_sku.feature`:

```gherkin
@api @catalog
Feature: Deactivate SKU
  As a catalog manager
  I want to deactivate SKUs
  So that they are no longer available for sale

  Background:
    Given the API server is running
    And the database is clean
    And a SKU exists with code "APPLE-001"

  Scenario: Deactivate an active SKU
    When I send a POST request to "/api/v1/skus/{sku_id}/deactivate"
    Then the response status should be 200
    And the response field "active" should be "false"
```

### Step 2: Run the test (it will fail with undefined steps)

```bash
make test-bdd-catalog
```

### Step 3: Implement the step definitions

The test output will show you which steps are undefined and provide snippets to implement them.

### Step 4: Implement the feature

Add the deactivate endpoint to your HTTP handler.

### Step 5: Run the test again

```bash
make test-bdd-catalog
```

The test should now pass! 🎉

## Common Commands

```bash
# Run all tests
make test-bdd

# Run with progress format (faster output)
make test-bdd-progress

# Generate JUnit report for CI
make test-bdd-junit

# Run specific feature file
cd server
go test -v ./test/... -godog.paths="../features/catalog/manage_skus.feature"

# Run specific scenario by line number
cd server
go test -v ./test/... -godog.paths="../features/catalog/manage_skus.feature:9"

# Skip error-handling tests
cd server
go test -v ./test/... -godog.tags="~@error-handling"
```

## Troubleshooting

### "connection refused" error

Make sure PostgreSQL is running:

```bash
# Check if running
docker ps | grep postgres-test

# Start if not running
docker start postgres-test
```

### "table does not exist" error

Migrations should run automatically. If they don't, check the database connection string:

```bash
# Set explicitly
export DATABASE_URL="postgres://vending:vending@localhost:5432/vending_test?sslmode=disable"
make test-bdd
```

### Tests are slow

Run only smoke tests for quick feedback:

```bash
make test-bdd-smoke
```

## Next Steps

1. Read the full [README.md](README.md) for detailed documentation
2. Explore the [BDD-Godog skill](.claude/skills/bdd-godog/SKILL.md) for patterns and best practices
3. Check out [practical examples](.claude/skills/bdd-godog/references/practical-examples.md)
4. Write tests for your new features!

## Tips

- **Start with smoke tests**: Tag critical scenarios with `@smoke` for quick validation
- **Use Background**: Put common setup steps in Background to avoid repetition
- **Keep scenarios focused**: One scenario = one behavior
- **Write from user perspective**: Use business language, not technical jargon
- **Run tests frequently**: BDD tests give you confidence to refactor

Happy testing! 🚀
