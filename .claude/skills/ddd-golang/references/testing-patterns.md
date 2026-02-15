# Testing Patterns for Hexagonal Architecture in Go

## Philosophy

Each architectural layer has a distinct testing strategy:

| Layer | Strategy | Tools | Speed |
|---|---|---|---|
| Domain | Pure unit tests | Standard `testing` package | Instant |
| Application | Unit tests with mocks | `gomock` or hand-rolled | Fast |
| Infrastructure | Integration tests | `testcontainers-go` | Slow (DB) |
| Interface Adapter | Handler tests | `net/http/httptest` | Fast |

The domain is the most important layer to test — it encodes your business rules. Because it has
no external dependencies, tests are fast, simple, and reliable.

---

## 1. Testing the Domain Layer

No mocks needed. Just construct aggregates and assert behaviour.

```go
// internal/domain/order/order_test.go
package order_test

import (
    "testing"

    "github.com/myorg/myservice/internal/domain/order"
)

func TestOrder_Confirm_SucceedsWithItems(t *testing.T) {
    o, err := order.NewOrder("customer-123")
    if err != nil {
        t.Fatalf("unexpected error creating order: %v", err)
    }

    price, _ := order.NewMoney(1000, "USD")
    if err := o.AddItem("prod-abc", 2, price); err != nil {
        t.Fatalf("unexpected error adding item: %v", err)
    }

    if err := o.Confirm(); err != nil {
        t.Errorf("expected confirm to succeed, got: %v", err)
    }

    if o.Status() != order.OrderStatusConfirmed {
        t.Errorf("expected status Confirmed, got %s", o.Status())
    }

    // Verify the domain event was emitted
    events := o.PullEvents()
    if len(events) != 1 {
        t.Fatalf("expected 1 event, got %d", len(events))
    }
    placed, ok := events[0].(order.OrderPlaced)
    if !ok {
        t.Fatalf("expected OrderPlaced event, got %T", events[0])
    }
    if placed.OrderID != o.ID() {
        t.Errorf("event OrderID mismatch")
    }
}

func TestOrder_Confirm_FailsWithNoItems(t *testing.T) {
    o, _ := order.NewOrder("customer-123")

    err := o.Confirm()

    if err == nil {
        t.Fatal("expected an error confirming an empty order")
    }
    if !errors.Is(err, order.ErrOrderHasNoItems) {
        t.Errorf("expected ErrOrderHasNoItems, got: %v", err)
    }
}

func TestMoney_Add_DifferentCurrencies_ReturnsError(t *testing.T) {
    usd, _ := order.NewMoney(100, "USD")
    eur, _ := order.NewMoney(100, "EUR")

    _, err := usd.Add(eur)

    if err == nil {
        t.Error("expected error when adding different currencies")
    }
}

// Table-driven test for VO construction validation
func TestNewMoney_Validation(t *testing.T) {
    cases := []struct {
        name     string
        amount   int64
        currency string
        wantErr  bool
    }{
        {"valid", 100, "USD", false},
        {"negative amount", -1, "USD", true},
        {"invalid currency code", 100, "US", true},
        {"zero amount is OK", 0, "USD", false},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            _, err := order.NewMoney(tc.amount, tc.currency)
            if (err != nil) != tc.wantErr {
                t.Errorf("NewMoney(%d, %q) error = %v, wantErr = %v", tc.amount, tc.currency, err, tc.wantErr)
            }
        })
    }
}
```

---

## 2. Testing the Application Layer (with Mocks)

Mock the repository and output port interfaces. The use case logic is tested in isolation.

### Option A: Hand-rolled mocks (simple, zero dependencies)

The simplest approach: implement the interface inline in your test file. No code generation needed.

```go
// internal/application/placeorder/mocks_test.go
package placeorder_test

import (
    "context"

    "github.com/myorg/myservice/internal/domain/order"
)

// mockOrderRepository is a hand-rolled in-memory mock.
type mockOrderRepository struct {
    saved  []*order.Order
    findFn func(ctx context.Context, id order.OrderID) (*order.Order, error)
}

func (m *mockOrderRepository) Save(_ context.Context, o *order.Order) error {
    m.saved = append(m.saved, o)
    return nil
}

func (m *mockOrderRepository) FindByID(ctx context.Context, id order.OrderID) (*order.Order, error) {
    if m.findFn != nil {
        return m.findFn(ctx, id)
    }
    return nil, order.ErrOrderNotFound
}

func (m *mockOrderRepository) FindByCustomer(_ context.Context, _ string) ([]*order.Order, error) {
    return nil, nil
}

// mockEventPublisher records what was published.
type mockEventPublisher struct {
    published []order.DomainEvent
}

func (m *mockEventPublisher) Publish(_ context.Context, evt order.DomainEvent) error {
    m.published = append(m.published, evt)
    return nil
}
```

For services with local interfaces ([threedots pattern](https://threedots.tech/post/introducing-clean-architecture/)),
mocks are even simpler — they only need to satisfy the minimal local interface:

```go
// internal/app/training_service_test.go
package app_test

import (
    "context"
    "time"
)

type trainerServiceMock struct {
    trainingsCancelled []time.Time
}

func (t *trainerServiceMock) CancelTraining(_ context.Context, trainingTime time.Time) error {
    t.trainingsCancelled = append(t.trainingsCancelled, trainingTime)
    return nil
}

type userServiceMock struct {
    balanceUpdates map[string]int
}

func (u *userServiceMock) UpdateTrainingBalance(_ context.Context, userUUID string, delta int) error {
    if u.balanceUpdates == nil {
        u.balanceUpdates = make(map[string]int)
    }
    u.balanceUpdates[userUUID] += delta
    return nil
}
```

This approach scales well and requires no external dependencies.
```

```go
// internal/application/placeorder/handler_test.go
package placeorder_test

import (
    "context"
    "testing"

    "github.com/myorg/myservice/internal/application/placeorder"
    "github.com/myorg/myservice/internal/domain/order"
)

func TestPlaceOrderHandler_Handle_Success(t *testing.T) {
    repo := &mockOrderRepository{}
    publisher := &mockEventPublisher{}
    handler := placeorder.NewPlaceOrderHandler(repo, publisher)

    cmd := placeorder.PlaceOrderCommand{
        CustomerID: "customer-123",
        Items: []placeorder.OrderItemInput{
            {ProductID: "prod-abc", Quantity: 1, UnitPriceCents: 5000, Currency: "USD"},
        },
    }

    result, err := handler.Handle(context.Background(), cmd)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.OrderID == "" {
        t.Error("expected a non-empty order ID")
    }
    if len(repo.saved) != 1 {
        t.Errorf("expected 1 saved order, got %d", len(repo.saved))
    }
    if len(publisher.published) != 1 {
        t.Errorf("expected 1 published event, got %d", len(publisher.published))
    }
    if _, ok := publisher.published[0].(order.OrderPlaced); !ok {
        t.Errorf("expected OrderPlaced event, got %T", publisher.published[0])
    }
}

func TestPlaceOrderHandler_Handle_EmptyCustomer_ReturnsError(t *testing.T) {
    handler := placeorder.NewPlaceOrderHandler(&mockOrderRepository{}, &mockEventPublisher{})

    _, err := handler.Handle(context.Background(), placeorder.PlaceOrderCommand{
        CustomerID: "", // invalid
        Items:      []placeorder.OrderItemInput{{ProductID: "p", Quantity: 1, UnitPriceCents: 100, Currency: "USD"}},
    })

    if err == nil {
        t.Error("expected error for empty customer ID")
    }
}
```

### Option B: gomock (better for complex interactions)

```go
// Generate mocks: mockgen -source=internal/domain/order/repository.go -destination=internal/domain/order/mocks/mock_repository.go

func TestPlaceOrderHandler_RepoSaveError(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockOrderRepository(ctrl)
    mockPub  := mocks.NewMockEventPublisher(ctrl)

    // Expect Save to be called once and return an error
    mockRepo.EXPECT().
        Save(gomock.Any(), gomock.Any()).
        Return(errors.New("db connection lost"))

    handler := placeorder.NewPlaceOrderHandler(mockRepo, mockPub)

    _, err := handler.Handle(context.Background(), placeorder.PlaceOrderCommand{
        CustomerID: "c1",
        Items: []placeorder.OrderItemInput{{ProductID: "p", Quantity: 1, UnitPriceCents: 100, Currency: "USD"}},
    })

    if err == nil {
        t.Error("expected error when repo.Save fails")
    }
}
```

---

## 3. Testing the Infrastructure Layer (Integration Tests)

Use `testcontainers-go` to spin up a real database. These tests are slower but verify that
SQL, migrations, and row scanning actually work.

```go
// internal/infrastructure/persistence/postgres/order_repo_integration_test.go
//go:build integration

package postgres_test

import (
    "context"
    "database/sql"
    "testing"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    _ "github.com/lib/pq"

    pgadapter "github.com/myorg/myservice/internal/infrastructure/persistence/postgres"
    "github.com/myorg/myservice/internal/domain/order"
)

func TestPostgresOrderRepository_SaveAndFind(t *testing.T) {
    ctx := context.Background()

    // Start a real Postgres instance via Docker
    pgContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:16-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    if err != nil {
        t.Fatalf("could not start postgres container: %v", err)
    }
    t.Cleanup(func() { pgContainer.Terminate(ctx) })

    connStr, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        t.Fatal(err)
    }

    // Run migrations (simplified — use golang-migrate in production)
    db.ExecContext(ctx, `CREATE TABLE orders (
        id TEXT PRIMARY KEY,
        customer_id TEXT NOT NULL,
        status TEXT NOT NULL,
        placed_at TIMESTAMPTZ
    )`)

    repo := pgadapter.NewPostgresOrderRepository(db)

    // Create and save an order
    o, _ := order.NewOrder("customer-123")
    price, _ := order.NewMoney(1000, "USD")
    o.AddItem("prod-abc", 1, price)
    o.Confirm()

    if err := repo.Save(ctx, o); err != nil {
        t.Fatalf("Save failed: %v", err)
    }

    // Retrieve it
    found, err := repo.FindByID(ctx, o.ID())
    if err != nil {
        t.Fatalf("FindByID failed: %v", err)
    }
    if found.ID() != o.ID() {
        t.Errorf("ID mismatch: got %s, want %s", found.ID(), o.ID())
    }
    if found.Status() != order.OrderStatusConfirmed {
        t.Errorf("status mismatch: got %s", found.Status())
    }
}

func TestPostgresOrderRepository_FindByID_NotFound(t *testing.T) {
    // ... (similar setup) ...
    _, err := repo.FindByID(ctx, order.OrderIDFrom("nonexistent-id"))
    if !errors.Is(err, order.ErrOrderNotFound) {
        t.Errorf("expected ErrOrderNotFound, got %v", err)
    }
}
```

---

## 4. Testing the HTTP Interface Adapter

```go
// internal/infrastructure/http/handlers/order_handler_test.go
package handlers_test

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/myorg/myservice/internal/application/placeorder"
    "github.com/myorg/myservice/internal/infrastructure/http/handlers"
)

// stubPlaceOrderHandler simulates the application use case for handler tests.
type stubPlaceOrderHandler struct {
    returnResult placeorder.PlaceOrderResult
    returnErr    error
}

func (s *stubPlaceOrderHandler) Handle(_ context.Context, _ placeorder.PlaceOrderCommand) (placeorder.PlaceOrderResult, error) {
    return s.returnResult, s.returnErr
}

func TestOrderHandler_PlaceOrder_Returns201(t *testing.T) {
    stub := &stubPlaceOrderHandler{
        returnResult: placeorder.PlaceOrderResult{OrderID: "order-xyz"},
    }
    handler := handlers.NewOrderHandler(stub)

    body := `{"customer_id":"c1","items":[{"product_id":"p1","quantity":1,"unit_price_cents":1000,"currency":"USD"}]}`
    req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    handler.PlaceOrder(rec, req)

    if rec.Code != http.StatusCreated {
        t.Errorf("expected status 201, got %d", rec.Code)
    }

    var resp map[string]string
    json.NewDecoder(rec.Body).Decode(&resp)
    if resp["order_id"] != "order-xyz" {
        t.Errorf("unexpected order_id in response: %v", resp)
    }
}

func TestOrderHandler_PlaceOrder_MalformedBody_Returns400(t *testing.T) {
    handler := handlers.NewOrderHandler(&stubPlaceOrderHandler{})

    req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(`not json`))
    rec := httptest.NewRecorder()

    handler.PlaceOrder(rec, req)

    if rec.Code != http.StatusBadRequest {
        t.Errorf("expected 400, got %d", rec.Code)
    }
}
```

---

## Running Tests by Layer

```bash
# Domain and application — fast, no external dependencies
go test ./internal/domain/... ./internal/application/...

# Infrastructure integration tests — require Docker
go test -tags=integration ./internal/infrastructure/...

# Interface adapter tests
go test ./internal/infrastructure/http/...

# All tests
go test ./...

# With race detector (always use in CI)
go test -race ./...
```
