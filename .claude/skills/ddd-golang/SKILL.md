---
name: ddd-golang
description: >
  Design and generate production-quality Go code following Hexagonal Architecture (Ports & Adapters)
  and Domain-Driven Design (DDD) tactical patterns. Use this skill whenever the user asks to scaffold
  a Go service, design a domain model, implement repositories, use cases, application services, domain
  events, value objects, or aggregates. Also trigger when the user says things like "clean architecture
  in Go", "DDD in Go", "ports and adapters", "domain layer", "bounded context", or asks to structure
  a Go microservice or module. Use this even when the user does not explicitly say DDD — if they're
  asking for well-structured Go services with clear separation of concerns, this skill applies.
---

# Golang Hexagonal Architecture & DDD Skill

This skill produces idiomatic Go code organized around **Hexagonal Architecture** (Ports & Adapters)
and **DDD tactical patterns**. The goal is clean separation between domain logic, application
orchestration, and infrastructure concerns — with zero circular dependencies and genuinely testable code.

---

## Mental Model

Think in three concentric circles:

```
┌─────────────────────────────────────────────┐
│  Infrastructure / Interface (outer ring)     │
│  ┌───────────────────────────────────────┐  │
│  │  Application (middle ring)             │  │
│  │  ┌─────────────────────────────────┐  │  │
│  │  │  Domain (inner core)             │  │  │
│  │  │  Entities, VOs, Aggregates,      │  │  │
│  │  │  Domain Events, Domain Services  │  │  │
│  │  └─────────────────────────────────┘  │  │
│  │  Use Cases, Ports (interfaces),        │  │
│  │  Application Services, DTOs            │  │
│  └───────────────────────────────────────┘  │
│  HTTP handlers, gRPC, DB repos, queues       │
└─────────────────────────────────────────────┘
```

**The cardinal rule**: dependencies always point inward. Domain knows nothing about application
or infrastructure. Application knows domain but not infrastructure. Infrastructure depends on both
but is depended on by nothing inside.

---

## Project Structure

Two valid structures are commonly used. Choose based on team preference and project size.

### Structure: Modular Bounded Contexts

For services with multiple bounded contexts, organize by context rather than layer:

```
my-service/
├── cmd/
│   └── server/
│       └── main.go                   # Wiring & bootstrap
├── internal/
│   ├── shared/                       # SHARED KERNEL
│   │   ├── valueobjects/             # Money, Weight, strongly-typed IDs
│   │   ├── events/                   # DomainEvent interface, BaseEvent
│   │   ├── policy/                   # Shared business policies
│   │   └── errors/                   # Shared domain errors
│   │
│   ├── catalog/                      # CATALOG BOUNDED CONTEXT
│   │   ├── domain/                   # SKU aggregate, repository port
│   │   │   ├── sku.go                # Aggregate root
│   │   │   ├── repository.go         # Repository interface
│   │   │   ├── events.go             # Domain events
│   │   │   └── errors.go             # Domain errors
│   │   ├── app/                      # Use cases (flat, one file per use case)
│   │   │   ├── create_sku.go         # CreateSKUHandler + Command + Result
│   │   │   └── queries.go            # SKUQueryService
│   │   ├── infra/                    # Infrastructure adapters
│   │   │   ├── postgres_repo.go      # Repository implementation
│   │   │   ├── http_handler.go       # HTTP handlers
│   │   │   └── routes.go             # Route registration
│   │   └── api/                      # PUBLIC API for other contexts
│   │       └── reader.go             # SKUReader interface + SKUView DTO
│   │
│   ├── device/                       # DEVICE BOUNDED CONTEXT
│   │   ├── domain/
│   │   ├── app/
│   │   ├── infra/
│   │   └── api/                      # DeviceReader interface + DeviceView DTO
│   │
│   ├── transaction/                  # TRANSACTION BOUNDED CONTEXT
│   │   ├── domain/                   # Session aggregate, DetectedItem VO
│   │   ├── app/
│   │   │   ├── ports/                # Cross-context port interfaces
│   │   │   │   ├── device_reader.go  # What we need from Device context
│   │   │   │   └── catalog_reader.go # What we need from Catalog context
│   │   │   ├── start_session.go
│   │   │   ├── submit_detection.go
│   │   │   └── queries.go
│   │   ├── infra/
│   │   │   ├── adapters/             # Implements app/ports using other APIs
│   │   │   │   ├── device_adapter.go
│   │   │   │   └── catalog_adapter.go
│   │   │   ├── postgres_repo.go
│   │   │   └── http_handler.go
│   │   └── api/                      # SessionReader interface
│   │
│   ├── platform/                     # SHARED INFRASTRUCTURE
│   │   ├── http/                     # Router (composes all context routes)
│   │   ├── postgres/                 # Migrations
│   │   └── messaging/                # Event publisher
│   │
│   └── pkg/                          # Shared utilities
│       └── logger/
└── go.mod
```

**Key principles for modular structure:**

- Each bounded context is self-contained with its own domain/app/infra layers
- Cross-context communication only through `api/` packages (never import another context's domain or infra)
- Consumer-defined ports: Transaction defines what it needs from Device/Catalog in `app/ports/`
- Adapters in `infra/adapters/` implement those ports using the other context's `api/` interfaces

### Cross-Context Communication Pattern

```
Transaction Context                    Device Context
┌─────────────────────────────┐       ┌─────────────────────────────┐
│ app/ports/device_reader.go  │       │ api/reader.go               │
│ ┌─────────────────────────┐ │       │ ┌─────────────────────────┐ │
│ │ type DeviceReader       │ │       │ │ type DeviceReader       │ │
│ │ interface {             │ │       │ │ interface {             │ │
│ │   FindByMachineID(...)  │ │       │ │   FindByMachineID(...)  │ │
│ │ }                       │ │       │ │ }                       │ │
│ └─────────────────────────┘ │       │ └─────────────────────────┘ │
│                             │       │                             │
│ infra/adapters/             │       │ + DeviceReaderAdapter       │
│ device_adapter.go           │──────>│   (implements interface)    │
│ (implements port using API) │       │                             │
└─────────────────────────────┘       └─────────────────────────────┘
```

**Wiring in main.go:**

```go
// Device context
deviceRepo := deviceinfra.NewPostgresDeviceRepository(pool)
deviceReader := deviceapi.NewDeviceReaderAdapter(deviceRepo)  // Exposes API

// Transaction context (uses Device's API via adapter)
deviceAdapter := transactionadapters.NewDeviceAdapter(deviceReader)
startSessionHandler := transactionapp.NewStartSessionHandler(deviceAdapter, ...)
```

---

## Core Patterns — Quick Reference

| Pattern | Where | Key Trait |
|---|---|---|
| **Entity** | `domain/` | Identity-based equality; mutable state only via methods |
| **Value Object** | `domain/` | Value-based equality; always immutable |
| **Aggregate Root** | `domain/` | Consistency boundary; sole entry point for mutations |
| **Repository Port** | `domain/` or `app/` | Interface defined by consumer, implemented by adapter |
| **Domain Event** | `domain/` | Immutable fact; past-tense name; collected by aggregate |
| **Domain Service** | `domain/` | Pure logic that spans multiple aggregates |
| **Use Case / App Service** | `application/` or `app/` | Orchestrates domain objects and output ports |
| **Output Port** | `application/ports/` or local | Interface for side-effects (events, email, SMS…) |
| **Adapter** | `infrastructure/` or `adapters/` | Implements port interfaces (DB, HTTP clients, queues) |
| **Entry Point (Port)** | `infrastructure/http` or `ports/` | Translates protocol → application command/query |
| **Cross-Context Port** | `<context>/app/ports/` | Interface for reading from other bounded contexts |
| **Cross-Context Adapter** | `<context>/infra/adapters/` | Implements cross-context port using another context's API |
| **Context API** | `<context>/api/` | Public read interface + DTOs exposed to other contexts |

**Ports vs Adapters clarification:**
- **Port** = interface (what you need) — defined by the consumer
- **Adapter** = implementation (how you get it) — implements the port
- In threedots terminology: "Ports" folder contains *entry points* (HTTP handlers), "Adapters" folder contains *external integrations* (DB repos, clients)
- For cross-context: Consumer context defines ports in `app/ports/`, implements them in `infra/adapters/` using provider's `api/`

Read `references/ddd-patterns.md` for complete annotated Go code for each pattern.
Read `references/testing-patterns.md` for layer-specific testing strategies and examples.

---

## Go-Specific Idioms to Always Follow

**Interfaces belong to the consumer.**
Define repository and output-port interfaces in the package that *uses* them (domain or application),
never in the infrastructure package that implements them. This is what keeps dependencies pointing inward.

Go's implicit interface satisfaction enables this naturally — adapters implement domain interfaces
without importing them, the consumer defines what it needs.

**Local interface definitions (threedots pattern).**
For application services, you can define interfaces *privately within the service file* rather than
in a shared `ports/` package. This keeps interfaces minimal and close to their usage:

```go
// internal/app/training_service.go

// Defined locally — only what this service needs
type trainingRepository interface {
    FindByID(ctx context.Context, id string) (*Training, error)
    Save(ctx context.Context, t *Training) error
}

type TrainingService struct {
    repo trainingRepository
}
```

This approach works well when interfaces are consumed by only one service. Use shared `ports/` packages
when multiple services need the same interface.

**Errors are values.**
Domain errors should be typed sentinel errors or structured error types. Return them explicitly.
Never panic for business rule violations.

**Constructors validate invariants.**
Write `NewOrder(...) (*Order, error)`. Reject invalid state at construction time so the rest of
the codebase can trust that any `Order` it receives is valid.

**Panic on nil dependencies in service constructors.**
For application services (not domain types), panic on nil required dependencies. This catches
wiring mistakes at startup rather than at runtime:

```go
func NewTrainingService(repo trainingRepository, users userService) *TrainingService {
    if repo == nil {
        panic("nil trainingRepository")
    }
    if users == nil {
        panic("nil userService")
    }
    return &TrainingService{repo: repo, users: users}
}
```

**Aggregates own their mutations.**
All state changes go through aggregate methods. No external code may set fields directly.
Exported fields on aggregates are a design smell.

**Value Objects are immutable.**
A VO method that "changes" a VO returns a new VO. Never mutate in place.

**Domain events are collected, not dispatched immediately.**
The aggregate root accumulates events in `[]DomainEvent`. The application service drains and
publishes them *after* the aggregate has been successfully persisted.

**Separate models per layer.**
Never share types across architectural boundaries. Each layer has its own model:

```go
// Adapter layer (DB-specific, has struct tags)
type TrainingModel struct {
    UUID     string    `db:"uuid"`
    UserUUID string    `db:"user_uuid"`
    Time     time.Time `db:"time"`
}

// Domain/Application layer (clean, no tags)
type Training struct {
    uuid     string
    userUUID string
    time     time.Time
}

func (t Training) CanBeCancelled() bool {
    return t.time.Sub(time.Now()) > 24*time.Hour
}
```

This separation allows layers to evolve independently — database schema changes don't cascade
through business logic, and domain types stay free of infrastructure concerns.

---

## Generating a New Feature — Step by Step

### 1. Model the Domain First
- Name the aggregate and list its invariants
- Identify Value Objects (anything identity-less: money, email, status, address)
- Define the repository port interface (just `Save`, `FindByID`, etc.)
- Name domain events that the aggregate should emit

### 2. Write the Application Use Case
- `Command` struct: flat input DTO, plain Go types only
- `Handler` struct: holds dependencies (repository, output ports) as interfaces
- `Handle(ctx, cmd)`: load → mutate → save → publish events
- Define output port interfaces in `application/ports/` if they don't exist yet

### 3. Implement the Infrastructure Adapters
- Repository adapter: implement the domain interface; map between DB rows and domain types here
- Output port adapters: Kafka publisher, HTTP client, SMTP sender, etc.
- Keep all SQL/ORM/HTTP client details inside the adapter; domain never sees them

### 4. Wire the Interface Adapter
- HTTP/gRPC handler reads the request, maps it to the `Command`, calls `handler.Handle`
- Maps `(result, error)` back to the protocol response
- Error translation (domain error → HTTP 422, not-found → 404) lives here, not in the domain

### 5. Test Each Layer Independently
- **Domain**: pure unit tests — no mocks needed, just construct and assert
- **Application**: mock the interfaces (gomock or hand-rolled); no DB, no network
- **Infrastructure**: integration tests using testcontainers-go
- **Interface adapter**: use `net/http/httptest` for HTTP handlers

---

## Naming Conventions

| Concept | Convention | Example |
|---|---|---|
| Aggregate root | noun | `Order`, `Customer` |
| Repository port | `<Noun>Repository` | `OrderRepository` |
| Use case handler | `<Verb><Noun>Handler` | `PlaceOrderHandler` |
| Command DTO | `<Verb><Noun>Command` | `PlaceOrderCommand` |
| Domain event | past tense | `OrderPlaced`, `PaymentFailed` |
| Infrastructure repo | `Postgres<Noun>Repository` | `PostgresOrderRepository` |
| Output port | `<Noun>Publisher/Notifier` | `EventPublisher`, `CustomerNotifier` |
| Cross-context port | `<Noun>Reader` | `DeviceReader`, `CatalogReader` |
| Cross-context adapter | `<Noun>Adapter` | `DeviceAdapter`, `CatalogAdapter` |
| Cross-context DTO | `<Noun>Info` or `<Noun>View` | `DeviceInfo`, `SKUView` |

---

## Common Pitfalls

- **Anemic domain model** — aggregates that are just structs with getters and all logic in the
  application service. Push behaviour down into the domain where it belongs.
- **Infrastructure bleed** — SQL tags, JSON tags, or ORM embedded types on domain structs.
  Use separate persistence models and explicit mapping functions.
- **Publishing events inside the aggregate** — aggregates collect events; application services publish.
- **Fat use cases** — a handler with 8 injected dependencies is doing too much. Split it.
- **Skipping the Repository interface** — always define the interface in the domain, even if you
  only have one implementation. It makes the domain testable without a database.
- **Direct cross-context imports** — never import another context's `domain/` or `infra/`. Use `api/` only.
- **Leaking domain types across contexts** — each context should define its own DTOs for cross-context reads.
- **Bidirectional context dependencies** — if A depends on B and B depends on A, refactor to extract shared concepts.

---

## Reference Files

- **`references/ddd-patterns.md`** — Full annotated Go code for every tactical pattern
- **`references/testing-patterns.md`** — Testing strategies with example code per layer

## External References

- [Three Dots Labs: Introducing Clean Architecture](https://threedots.tech/post/introducing-clean-architecture/) — Alternative flat structure, local interface definitions
- [Three Dots Labs: Wild Workouts](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example) — Complete example repository
