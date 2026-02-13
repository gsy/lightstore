# DDD Tactical Patterns — Annotated Go Code

## Table of Contents
1. [Value Object](#1-value-object)
2. [Entity](#2-entity)
3. [Aggregate Root](#3-aggregate-root)
4. [Domain Events](#4-domain-events)
5. [Repository Port](#5-repository-port)
6. [Domain Service](#6-domain-service)
7. [Domain Errors](#7-domain-errors)
8. [Application Use Case](#8-application-use-case)
9. [Output Port](#9-output-port)
10. [Repository Adapter (Infrastructure)](#10-repository-adapter-infrastructure)
11. [Interface Adapter (HTTP)](#11-interface-adapter-http)
12. [Wiring in main.go](#12-wiring-in-maingo)

---

## 1. Value Object

Value Objects have no identity — two VOs with the same data are interchangeable. They are always
immutable. Any "modification" returns a new VO.

```go
// internal/domain/order/value_objects.go
package order

import (
    "errors"
    "fmt"
)

// Money is a Value Object. It has no ID.
// It's defined by its value (Amount + Currency).
type Money struct {
    amount   int64  // stored in minor units (cents) to avoid float precision issues
    currency string // ISO 4217 code e.g. "USD"
}

// NewMoney is the constructor — it validates and returns an error rather than a zero value.
func NewMoney(amount int64, currency string) (Money, error) {
    if len(currency) != 3 {
        return Money{}, errors.New("currency must be a 3-letter ISO 4217 code")
    }
    if amount < 0 {
        return Money{}, errors.New("money amount cannot be negative")
    }
    return Money{amount: amount, currency: currency}, nil
}

// Exported getters — the fields themselves are unexported to enforce immutability.
func (m Money) Amount() int64    { return m.amount }
func (m Money) Currency() string { return m.currency }

// Equals implements value-based equality.
func (m Money) Equals(other Money) bool {
    return m.amount == other.amount && m.currency == other.currency
}

// Add returns a NEW Money — it does not mutate m.
func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, fmt.Errorf("cannot add %s to %s", other.currency, m.currency)
    }
    return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

func (m Money) String() string {
    return fmt.Sprintf("%d %s", m.amount, m.currency)
}

// ---- Strongly-typed IDs (another common VO pattern) ----

// OrderID prevents accidentally mixing up IDs of different types.
type OrderID string

func NewOrderID() OrderID {
    return OrderID(generateUUID()) // generateUUID is a small helper wrapping google/uuid
}

func OrderIDFrom(raw string) (OrderID, error) {
    if raw == "" {
        return "", errors.New("order ID cannot be empty")
    }
    return OrderID(raw), nil
}

// ---- Status as a constrained VO ----

type OrderStatus string

const (
    OrderStatusDraft     OrderStatus = "draft"
    OrderStatusConfirmed OrderStatus = "confirmed"
    OrderStatusShipped   OrderStatus = "shipped"
    OrderStatusCancelled OrderStatus = "cancelled"
)

func (s OrderStatus) IsTerminal() bool {
    return s == OrderStatusShipped || s == OrderStatusCancelled
}
```

---

## 2. Entity

Entities have a stable identity. Two entities with different state are still the "same" entity
if they share the same ID. Entities may be mutable, but only through methods that enforce invariants.

```go
// internal/domain/order/entity.go
package order

import "errors"

// OrderItem is an Entity within the Order aggregate (not the root).
// It has its own identity (ItemID) but its lifecycle is governed by Order.
type OrderItem struct {
    id        OrderItemID
    productID string
    quantity  int
    unitPrice Money
}

func NewOrderItem(productID string, quantity int, unitPrice Money) (OrderItem, error) {
    if productID == "" {
        return OrderItem{}, errors.New("product ID is required")
    }
    if quantity <= 0 {
        return OrderItem{}, errors.New("quantity must be positive")
    }
    return OrderItem{
        id:        NewOrderItemID(),
        productID: productID,
        quantity:  quantity,
        unitPrice: unitPrice,
    }, nil
}

func (i OrderItem) ID() OrderItemID    { return i.id }
func (i OrderItem) ProductID() string  { return i.productID }
func (i OrderItem) Quantity() int      { return i.quantity }
func (i OrderItem) UnitPrice() Money   { return i.unitPrice }

func (i OrderItem) Subtotal() (Money, error) {
    total := i.unitPrice
    for j := 1; j < i.quantity; j++ {
        var err error
        total, err = total.Add(i.unitPrice)
        if err != nil {
            return Money{}, err
        }
    }
    return total, nil
}
```

---

## 3. Aggregate Root

The aggregate root is the entry point for all mutations within its consistency boundary. External
code interacts only with the root, never directly with inner entities. The root enforces all
invariants and collects domain events.

```go
// internal/domain/order/aggregate.go
package order

import (
    "errors"
    "time"
)

// Order is the aggregate root. It owns OrderItems.
// All business operations go through Order methods.
type Order struct {
    id         OrderID
    customerID string
    items      []OrderItem
    status     OrderStatus
    placedAt   *time.Time

    // Domain events are collected here and published by the application service after persistence.
    events []DomainEvent
}

// NewOrder is the factory for creating a brand-new order.
// It validates all inputs before producing a valid Order.
func NewOrder(customerID string) (*Order, error) {
    if customerID == "" {
        return nil, errors.New("customer ID is required")
    }
    o := &Order{
        id:         NewOrderID(),
        customerID: customerID,
        items:      []OrderItem{},
        status:     OrderStatusDraft,
    }
    return o, nil
}

// Reconstitute rebuilds an Order from persisted data (called by the repository adapter).
// No events are emitted; no invariant creation logic runs — the data is trusted.
func Reconstitute(id OrderID, customerID string, items []OrderItem, status OrderStatus, placedAt *time.Time) *Order {
    return &Order{
        id:         id,
        customerID: customerID,
        items:      items,
        status:     status,
        placedAt:   placedAt,
    }
}

// ---- Read accessors ----

func (o *Order) ID() OrderID         { return o.id }
func (o *Order) CustomerID() string  { return o.customerID }
func (o *Order) Items() []OrderItem  { return append([]OrderItem(nil), o.items...) } // defensive copy
func (o *Order) Status() OrderStatus { return o.status }

// ---- Business operations (invariant-enforcing mutations) ----

// AddItem appends a new item to the draft order.
func (o *Order) AddItem(productID string, quantity int, unitPrice Money) error {
    if o.status != OrderStatusDraft {
        return ErrOrderNotEditable
    }
    item, err := NewOrderItem(productID, quantity, unitPrice)
    if err != nil {
        return err
    }
    o.items = append(o.items, item)
    return nil
}

// Confirm transitions the order from draft to confirmed and emits an OrderPlaced event.
func (o *Order) Confirm() error {
    if o.status != OrderStatusDraft {
        return ErrOrderAlreadyConfirmed
    }
    if len(o.items) == 0 {
        return ErrOrderHasNoItems
    }

    now := time.Now().UTC()
    o.status = OrderStatusConfirmed
    o.placedAt = &now

    // Emit a domain event — do NOT call the event publisher here.
    // The application service will drain and publish after Save().
    o.events = append(o.events, OrderPlaced{
        OrderID:    o.id,
        CustomerID: o.customerID,
        PlacedAt:   now,
    })
    return nil
}

// Cancel transitions the order to cancelled if allowed.
func (o *Order) Cancel(reason string) error {
    if o.status.IsTerminal() {
        return ErrOrderCannotBeCancelled
    }
    o.status = OrderStatusCancelled
    o.events = append(o.events, OrderCancelled{
        OrderID: o.id,
        Reason:  reason,
    })
    return nil
}

// ---- Domain event drainage ----

// PullEvents returns all accumulated domain events and clears the internal slice.
// Called once by the application service after a successful Save().
func (o *Order) PullEvents() []DomainEvent {
    evts := o.events
    o.events = nil
    return evts
}
```

---

## 4. Domain Events

Domain events are immutable records of something that happened inside the domain. They use past-tense
names and carry only the data relevant to the event.

```go
// internal/domain/order/events.go
package order

import "time"

// DomainEvent is the marker interface for all domain events.
// It has no methods — it's used for type-safe slices only.
type DomainEvent interface {
    domainEvent()
}

// OrderPlaced fires when a draft order is confirmed by the customer.
type OrderPlaced struct {
    OrderID    OrderID
    CustomerID string
    PlacedAt   time.Time
}

func (OrderPlaced) domainEvent() {}

// OrderCancelled fires when an order is cancelled for any reason.
type OrderCancelled struct {
    OrderID OrderID
    Reason  string
}

func (OrderCancelled) domainEvent() {}
```

---

## 5. Repository Port

The repository interface is defined in the **domain** package. This is critical — it means the
domain has no import of the infrastructure layer. The interface expresses what the domain *needs*,
not how storage works.

```go
// internal/domain/order/repository.go
package order

import "context"

// OrderRepository is a PORT — an interface defined by and for the domain.
// The infrastructure layer implements this interface; the domain never imports infra.
type OrderRepository interface {
    // Save persists a new or updated order.
    Save(ctx context.Context, order *Order) error
    // FindByID returns ErrOrderNotFound if the order does not exist.
    FindByID(ctx context.Context, id OrderID) (*Order, error)
    // FindByCustomer returns all orders for a given customer.
    FindByCustomer(ctx context.Context, customerID string) ([]*Order, error)
}
```

---

## 6. Domain Service

Use a domain service when a piece of business logic is meaningful to the domain but doesn't
naturally belong to a single aggregate. Domain services are stateless and depend only on
domain types (no infrastructure).

```go
// internal/domain/order/service.go
package order

import "context"

// PricingService calculates order totals taking discounts into account.
// It lives in the domain because pricing rules ARE business rules.
// It does NOT call a database or HTTP endpoint — those would belong in the application layer.
type PricingService struct{}

type PricingResult struct {
    Subtotal Money
    Discount Money
    Total    Money
}

func (s *PricingService) CalculateTotal(order *Order, discountPercent int) (PricingResult, error) {
    var subtotal Money
    var err error

    for _, item := range order.Items() {
        sub, e := item.Subtotal()
        if e != nil {
            return PricingResult{}, e
        }
        subtotal, err = subtotal.Add(sub)
        if err != nil {
            return PricingResult{}, err
        }
    }

    discountAmount := Money{amount: subtotal.amount * int64(discountPercent) / 100, currency: subtotal.currency}
    total, err := subtotal.Add(Money{amount: -discountAmount.amount, currency: discountAmount.currency})
    if err != nil {
        return PricingResult{}, err
    }

    return PricingResult{Subtotal: subtotal, Discount: discountAmount, Total: total}, nil
}
```

---

## 7. Domain Errors

Domain errors should be typed so callers (application and interface adapters) can distinguish
them without string matching.

```go
// internal/domain/order/errors.go
package order

import "errors"

var (
    ErrOrderNotFound        = errors.New("order not found")
    ErrOrderNotEditable     = errors.New("order cannot be edited in its current status")
    ErrOrderAlreadyConfirmed = errors.New("order is already confirmed")
    ErrOrderHasNoItems      = errors.New("order must have at least one item before confirming")
    ErrOrderCannotBeCancelled = errors.New("order cannot be cancelled from its current status")
)
```

---

## 8. Application Use Case

The application service orchestrates: load from repo → call domain method → save → publish events.
It depends on domain types and port interfaces only — never on infrastructure concretions.

### Style: Service with Local Interfaces (threedots pattern)

From [Three Dots Labs](https://threedots.tech/post/introducing-clean-architecture/). Interfaces
are defined *locally* in the service file — only what this service needs, nothing more.

```go
// internal/app/training_service.go
package app

import (
    "context"
    "time"

    "github.com/pkg/errors"
)

// Training is the application-layer model. Clean of infrastructure tags.
// Business methods live here, not in external handlers.
type Training struct {
    UUID     string
    UserUUID string
    Time     time.Time
    Notes    string
}

func (t Training) CanBeCancelled() bool {
    return t.Time.Sub(time.Now()) > 24*time.Hour
}

// Interfaces defined locally — each service specifies exactly what it needs.
// Go's implicit satisfaction means adapters implement this without importing it.

type trainingRepository interface {
    CancelTraining(ctx context.Context, trainingUUID string,
        cancelFn func(t Training) error) error
}

type trainerService interface {
    CancelTraining(ctx context.Context, trainingTime time.Time) error
}

type userService interface {
    UpdateTrainingBalance(ctx context.Context, userUUID string, delta int) error
}

// TrainingService orchestrates training operations.
type TrainingService struct {
    repo     trainingRepository
    trainers trainerService
    users    userService
}

// NewTrainingService constructs the service. Panic on nil deps to catch wiring bugs at startup.
func NewTrainingService(
    repo trainingRepository,
    trainers trainerService,
    users userService,
) TrainingService {
    if repo == nil {
        panic("nil trainingRepository")
    }
    if trainers == nil {
        panic("nil trainerService")
    }
    if users == nil {
        panic("nil userService")
    }
    return TrainingService{repo: repo, trainers: trainers, users: users}
}

// CancelTraining cancels a training and adjusts balances.
// Business logic is in the service — the handler knows nothing about HTTP or storage.
func (s TrainingService) CancelTraining(ctx context.Context, user User, trainingUUID string) error {
    return s.repo.CancelTraining(ctx, trainingUUID, func(training Training) error {
        // Authorization check
        if user.Role != "trainer" && training.UserUUID != user.UUID {
            return errors.Errorf("user %q cannot cancel this training", user.UUID)
        }

        // Business rule: calculate balance delta based on cancellation time
        var balanceDelta int
        if training.CanBeCancelled() {
            balanceDelta = 1 // Refund the credit
        }

        if balanceDelta != 0 {
            if err := s.users.UpdateTrainingBalance(ctx, training.UserUUID, balanceDelta); err != nil {
                return errors.Wrap(err, "updating balance")
            }
        }

        return s.trainers.CancelTraining(ctx, training.Time)
    })
}
```
---

## 9. Output Port

Output ports are interfaces defined in the application layer for side-effects the use case
needs to trigger (publish events, send emails, call external APIs).

```go
// internal/application/ports/event_publisher.go
package ports

import (
    "context"

    "github.com/myorg/myservice/internal/domain/order"
)

// EventPublisher is an OUTPUT PORT — the application layer defines what it needs,
// the infrastructure layer provides a concrete implementation.
type EventPublisher interface {
    Publish(ctx context.Context, event order.DomainEvent) error
}
```

---

## 10. Repository Adapter (Infrastructure)

The adapter implements the domain repository interface using a specific technology (Postgres here).
All DB-specific details (SQL, rows, scanning) stay inside this file. It maps between DB rows and
domain objects.

```go
// internal/infrastructure/persistence/postgres/order_repo.go
package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "time"

    "github.com/myorg/myservice/internal/domain/order"
)

// PostgresOrderRepository implements order.OrderRepository.
// It is the ADAPTER for the repository port.
type PostgresOrderRepository struct {
    db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
    return &PostgresOrderRepository{db: db}
}

// orderRow is a DB-layer struct — it has db tags and SQL-friendly types.
// It never leaves this file. Domain types never contain these tags.
type orderRow struct {
    ID         string    `db:"id"`
    CustomerID string    `db:"customer_id"`
    Status     string    `db:"status"`
    PlacedAt   *time.Time `db:"placed_at"`
}

func (r *PostgresOrderRepository) Save(ctx context.Context, o *order.Order) error {
    _, err := r.db.ExecContext(ctx, `
        INSERT INTO orders (id, customer_id, status, placed_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO UPDATE
          SET status = EXCLUDED.status,
              placed_at = EXCLUDED.placed_at
    `, string(o.ID()), o.CustomerID(), string(o.Status()), nil)
    return err
}

func (r *PostgresOrderRepository) FindByID(ctx context.Context, id order.OrderID) (*order.Order, error) {
    row := r.db.QueryRowContext(ctx, `
        SELECT id, customer_id, status, placed_at FROM orders WHERE id = $1
    `, string(id))

    var rec orderRow
    if err := row.Scan(&rec.ID, &rec.CustomerID, &rec.Status, &rec.PlacedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, order.ErrOrderNotFound
        }
        return nil, fmt.Errorf("scanning order row: %w", err)
    }

    // Map from DB record to domain type using Reconstitute (no event emission, no validation)
    oid, _ := order.OrderIDFrom(rec.ID)
    return order.Reconstitute(
        oid,
        rec.CustomerID,
        nil, // items would be loaded in a separate query
        order.OrderStatus(rec.Status),
        rec.PlacedAt,
    ), nil
}

func (r *PostgresOrderRepository) FindByCustomer(ctx context.Context, customerID string) ([]*order.Order, error) {
    // Implementation omitted for brevity — same pattern as FindByID but with rows.Next()
    return nil, nil
}
```

---

## 11. Interface Adapter (HTTP)

The HTTP handler sits in the outermost ring. It translates HTTP requests into application
commands, calls the use case, and maps results (including errors) back to HTTP responses.
Domain error translation lives here, not in the domain or application layer.

```go
// internal/infrastructure/http/handlers/order_handler.go
package handlers

import (
    "encoding/json"
    "errors"
    "net/http"

    "github.com/myorg/myservice/internal/application/placeorder"
    "github.com/myorg/myservice/internal/domain/order"
)

type OrderHandler struct {
    placeOrder *placeorder.PlaceOrderHandler
}

func NewOrderHandler(placeOrder *placeorder.PlaceOrderHandler) *OrderHandler {
    return &OrderHandler{placeOrder: placeOrder}
}

type placeOrderRequest struct {
    CustomerID string          `json:"customer_id"`
    Items      []orderItemJSON `json:"items"`
}

type orderItemJSON struct {
    ProductID      string `json:"product_id"`
    Quantity       int    `json:"quantity"`
    UnitPriceCents int64  `json:"unit_price_cents"`
    Currency       string `json:"currency"`
}

type placeOrderResponse struct {
    OrderID string `json:"order_id"`
}

func (h *OrderHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
    var req placeOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    // Map HTTP request → application command (primitive types only)
    cmd := placeorder.PlaceOrderCommand{
        CustomerID: req.CustomerID,
    }
    for _, item := range req.Items {
        cmd.Items = append(cmd.Items, placeorder.OrderItemInput{
            ProductID:      item.ProductID,
            Quantity:       item.Quantity,
            UnitPriceCents: item.UnitPriceCents,
            Currency:       item.Currency,
        })
    }

    result, err := h.placeOrder.Handle(r.Context(), cmd)
    if err != nil {
        // Map domain errors → HTTP status codes HERE, not in the domain or application layer.
        switch {
        case errors.Is(err, order.ErrOrderHasNoItems):
            http.Error(w, err.Error(), http.StatusUnprocessableEntity)
        case errors.Is(err, order.ErrOrderNotFound):
            http.Error(w, err.Error(), http.StatusNotFound)
        default:
            http.Error(w, "internal server error", http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(placeOrderResponse{OrderID: result.OrderID})
}
```

---

## 12. Wiring in main.go

All dependency construction happens in `main.go` (or a dedicated `wire.go` / DI container).
This is the only place where concrete infrastructure types are instantiated and injected.

```go
// cmd/server/main.go
package main

import (
    "database/sql"
    "log"
    "net/http"

    _ "github.com/lib/pq"

    "github.com/myorg/myservice/internal/application/placeorder"
    "github.com/myorg/myservice/internal/infrastructure/http/handlers"
    "github.com/myorg/myservice/internal/infrastructure/messaging/kafka"
    "github.com/myorg/myservice/internal/infrastructure/persistence/postgres"
)

func main() {
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/mydb?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Infrastructure adapters (implement domain interfaces)
    orderRepo := postgres.NewPostgresOrderRepository(db)
    publisher := kafka.NewKafkaEventPublisher("localhost:9092")

    // Application use cases (receive interfaces, know nothing about infra concretions)
    placeOrderHandler := placeorder.NewPlaceOrderHandler(orderRepo, publisher)

    // Interface adapters (receive application use cases)
    orderHandler := handlers.NewOrderHandler(placeOrderHandler)

    mux := http.NewServeMux()
    mux.HandleFunc("POST /orders", orderHandler.PlaceOrder)

    log.Println("listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```
