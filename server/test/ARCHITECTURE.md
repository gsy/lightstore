# BDD Test Architecture

## Overview

This document describes the architecture of the BDD testing framework for the vending machine backend.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Feature Files (Gherkin)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   Catalog    │  │    Device    │  │    Transaction       │  │
│  │  Features    │  │   Features   │  │     Features         │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│                    Step Definitions (Go)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐  │
│  │   Common     │  │   Catalog    │  │    Device            │  │
│  │   Steps      │  │   Steps      │  │    Steps             │  │
│  └──────────────┘  └──────────────┘  └──────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Transaction Steps                            │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│                      Test Context                                │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  • HTTP Client                                             │ │
│  │  • Test Server (httptest)                                  │ │
│  │  • Database Pool                                           │ │
│  │  • Request/Response State                                  │ │
│  │  • Created Resources (SKUs, Devices, Sessions)            │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│                   Application Under Test                         │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                  HTTP Router (Gin)                         │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐ │ │
│  │  │   Catalog    │  │    Device    │  │  Transaction    │ │ │
│  │  │   Handler    │  │   Handler    │  │    Handler      │ │ │
│  │  └──────────────┘  └──────────────┘  └─────────────────┘ │ │
│  └────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │              Application Layer (Use Cases)                 │ │
│  │  • CreateSKUHandler      • RegisterDeviceHandler          │ │
│  │  • StartSessionHandler   • SubmitDetectionHandler         │ │
│  │  • ConfirmSessionHandler • CancelSessionHandler           │ │
│  └────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │                   Domain Layer                             │ │
│  │  • SKU Aggregate         • Device Aggregate               │ │
│  │  • Session Aggregate     • Value Objects                  │ │
│  │  • Domain Events         • Business Rules                 │ │
│  └────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │              Infrastructure Layer                          │ │
│  │  • PostgreSQL Repositories                                │ │
│  │  • Event Publisher (NoOp for tests)                       │ │
│  │  • Cross-Context Adapters                                 │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│                    Test Database                                 │
│                   PostgreSQL (vending_test)                      │
└─────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Feature Files Layer

**Purpose**: Define system behavior in business language

**Components**:
- `features/catalog/manage_skus.feature`
- `features/device/device_management.feature`
- `features/transaction/shopping_session.feature`

**Characteristics**:
- Written in Gherkin (Given-When-Then)
- Human-readable specifications
- Living documentation
- Tagged for organization (@smoke, @api, etc.)

### 2. Step Definitions Layer

**Purpose**: Map Gherkin steps to Go code

**Components**:
- `common_steps.go` - Shared steps (API calls, assertions)
- `catalog_steps.go` - Catalog-specific steps
- `device_steps.go` - Device-specific steps
- `transaction_steps.go` - Transaction-specific steps

**Characteristics**:
- Regex-based step matching
- Reusable across scenarios
- Interact with Test Context
- No direct business logic

### 3. Test Context Layer

**Purpose**: Manage test state and provide utilities

**Components**:
- `support/test_context.go` - State management
- `support/test_server.go` - Server setup

**Responsibilities**:
- HTTP client management
- Request/response storage
- Resource tracking (created SKUs, devices, sessions)
- Database cleanup
- Placeholder replacement

**State Management**:
```go
type TestContext struct {
    Server          *httptest.Server
    Client          *http.Client
    DBPool          *pgxpool.Pool
    LastRequest     *http.Request
    LastResponse    *http.Response
    LastBody        []byte
    CreatedSKUs     map[string]string
    CreatedDevices  map[string]string
    CreatedSessions map[string]string
}
```

### 4. Application Layer

**Purpose**: The actual system being tested

**Components**:
- HTTP Handlers (Gin)
- Application Services (Use Cases)
- Domain Layer (Aggregates, Value Objects)
- Infrastructure (Repositories, Adapters)

**Characteristics**:
- Real application code
- Full dependency injection
- Hexagonal Architecture
- DDD patterns

### 5. Database Layer

**Purpose**: Persistent storage for tests

**Components**:
- PostgreSQL container (postgres-test)
- Test database (vending_test)
- Migrations (run automatically)

**Characteristics**:
- Isolated from production
- Clean state per scenario
- Real database operations

## Test Execution Flow

### Scenario Lifecycle

```
1. Before Scenario Hook
   ├─ Reset test context
   └─ Clear request/response state

2. Background Steps
   ├─ Start test server
   └─ Clean database

3. Scenario Steps
   ├─ Given: Setup preconditions
   ├─ When: Perform actions
   └─ Then: Assert outcomes

4. After Scenario Hook
   ├─ Close test server
   └─ Cleanup resources
```

### Request Flow

```
Step Definition
    ↓
TestContext.SendRequest()
    ↓
HTTP Client
    ↓
Test Server (httptest)
    ↓
Gin Router
    ↓
HTTP Handler
    ↓
Application Service
    ↓
Domain Aggregate
    ↓
Repository
    ↓
Database
    ↓
Response
    ↓
TestContext (store response)
    ↓
Assertion Steps
```

## Data Flow

### Creating a Resource

```
Feature File:
  When I create a SKU with code "APPLE-001"

Step Definition:
  func iCreateSKU(code string) {
      sku := map[string]interface{}{
          "code": code,
          "name": "Test Product",
          ...
      }
      testContext.SendRequest("POST", "/api/v1/skus", sku)
  }

Test Context:
  func SendRequest(method, path, body) {
      req := http.NewRequest(method, server.URL + path, body)
      resp := client.Do(req)
      lastResponse = resp
      lastBody = readBody(resp)
  }

Application:
  POST /api/v1/skus
    → HTTPHandler.Create()
    → CreateSKUHandler.Handle()
    → SKU.New() (domain)
    → repository.Save()
    → PostgreSQL INSERT

Response:
  201 Created
  { "id": "uuid", "message": "SKU created" }

Test Context:
  Store SKU ID: createdSKUs["APPLE-001"] = "uuid"

Assertion:
  Then the response status should be 201
    → assert lastResponse.StatusCode == 201
```

## Isolation Strategy

### Scenario Isolation

Each scenario is completely isolated:

1. **Database**: Truncated before each scenario
2. **Server**: New httptest.Server instance
3. **State**: Fresh TestContext
4. **Resources**: Tracked and cleaned up

### No Shared State

```
Scenario 1                 Scenario 2
    ↓                          ↓
Clean DB                   Clean DB
    ↓                          ↓
New Server                 New Server
    ↓                          ↓
Fresh Context              Fresh Context
    ↓                          ↓
Run Steps                  Run Steps
    ↓                          ↓
Cleanup                    Cleanup
```

## Cross-Context Testing

### Transaction Context Dependencies

```
Transaction Context
    ↓
Needs Device Info
    ↓
DeviceAdapter (infra/adapters)
    ↓
DeviceReader (api)
    ↓
Device Context

Transaction Context
    ↓
Needs SKU Info
    ↓
CatalogAdapter (infra/adapters)
    ↓
SKUReader (api)
    ↓
Catalog Context
```

### Test Setup for Cross-Context

```gherkin
Background:
  Given the API server is running
  And the database is clean
  And a device exists with machine ID "DEVICE-001"
  And the following SKUs exist:
    | code      | name       |
    | APPLE-001 | Fuji Apple |
```

This creates:
1. Device in device context
2. SKUs in catalog context
3. Transaction can now use both

## Test Patterns

### Pattern 1: Resource Creation

```
Given a resource exists
    ↓
Create via API
    ↓
Store ID in TestContext
    ↓
Use in subsequent steps
```

### Pattern 2: State Verification

```
When action is performed
    ↓
Store response
    ↓
Then verify response
    ↓
Assert status, fields, values
```

### Pattern 3: Error Handling

```
When invalid action
    ↓
Expect error response
    ↓
Verify error status
    ↓
Verify error message
```

### Pattern 4: Placeholder Replacement

```
Feature: Get SKU by ID
  When I send a GET request to "/api/v1/skus/{sku_id}"
                                                  ↑
                                    Replaced with actual ID
                                    from TestContext.CreatedSKUs
```

## Performance Considerations

### Test Speed

- **Smoke tests**: ~2-5 seconds (critical paths only)
- **Full suite**: ~10-30 seconds (all scenarios)
- **Per scenario**: ~0.5-2 seconds

### Optimization Strategies

1. **Use Background**: Common setup runs once per scenario
2. **Tag organization**: Run subsets with tags
3. **Parallel execution**: Godog supports parallel scenarios
4. **In-memory adapters**: Could replace DB for faster tests

### Trade-offs

- **Speed vs Reality**: Real DB is slower but more accurate
- **Isolation vs Performance**: Clean state per scenario is slower but safer
- **Coverage vs Time**: More scenarios = more time

## Extension Points

### Adding New Context

1. Create feature file: `features/newcontext/feature.feature`
2. Create step definitions: `test/newcontext_steps.go`
3. Register steps in `bdd_test.go`
4. Add tag: `@newcontext`
5. Add Makefile target: `test-bdd-newcontext`

### Adding New Scenarios

1. Add to existing feature file
2. Reuse existing steps when possible
3. Add new steps if needed
4. Tag appropriately

### Custom Assertions

```go
func theResponseShouldContainValidUUID(field string) error {
    value, err := testContext.GetNestedField(field)
    if err != nil {
        return err
    }
    
    if _, err := uuid.Parse(fmt.Sprint(value)); err != nil {
        return fmt.Errorf("field %s is not a valid UUID", field)
    }
    
    return nil
}
```

## Debugging

### View Request/Response

```go
func debugLastRequest() {
    fmt.Printf("Request: %s %s\n", 
        testContext.LastRequest.Method,
        testContext.LastRequest.URL)
    fmt.Printf("Response: %d\n%s\n",
        testContext.LastResponse.StatusCode,
        string(testContext.LastBody))
}
```

### Database State

```bash
# Connect to test database
docker exec -it postgres-test psql -U vending -d vending_test

# View tables
\dt

# Query data
SELECT * FROM catalog_skus;
```

### Test Output

```bash
# Verbose output
go test -v ./test/...

# With colors
go test -v ./test/... -godog.format=pretty -godog.no-colors=false

# Stop on first failure
go test -v ./test/... -godog.stop-on-failure
```

## Best Practices

1. **Keep scenarios independent**: No shared state
2. **Use Background wisely**: Common setup only
3. **Tag appropriately**: Organize with tags
4. **Reuse steps**: DRY principle
5. **Clean state**: Always start fresh
6. **Test real paths**: Use actual application code
7. **Document patterns**: Update this file

## Conclusion

This architecture provides:
- ✅ Clear separation of concerns
- ✅ Reusable components
- ✅ Isolated test scenarios
- ✅ Real integration testing
- ✅ Easy to extend
- ✅ Fast feedback loop
- ✅ Living documentation

The BDD tests serve as both executable specifications and regression tests, giving confidence to refactor and extend the system.
