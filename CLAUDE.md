# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Beverage vending machine recognition system using ESP32-S3-CAM for computer vision and HX711 load cell for weight verification. Hybrid ML approach: on-device TFLite inference with cloud fallback.

## Commands

### Backend Server (Go)

```bash
# Start with Docker (recommended)
make docker-up          # Start postgres + server
make docker-down        # Stop all services
make docker-logs        # View server logs

# Local development
make deps               # Download Go dependencies
make build              # Build binary to bin/server
make run                # Run server (requires postgres)
make test               # Run tests

# Database
make db-connect         # Connect to postgres via psql
make db-seed            # Seed sample beverages

# Kubernetes
make k8s-deploy         # Deploy to k8s cluster
make k8s-delete         # Remove from k8s
```

### Firmware (ESP32-S3)

```bash
cd firmware
pio run                 # Build firmware
pio run -t upload       # Flash to device
pio device monitor      # Serial monitor (115200 baud)
```

## Architecture

### Three-Tier System

1. **ESP32-S3-CAM Device** (`firmware/`) - Captures images, measures weight, runs on-device ML inference
2. **Backend Server** (`server/`) - Go/Gin REST API, PostgreSQL, DDD hexagonal architecture
3. **Mobile App** (external) - Customer-facing app for QR scanning, payment, refunds

### Backend DDD Modular Architecture

The backend follows Domain-Driven Design with modular bounded contexts. Each context is self-contained with its own domain, application, and infrastructure layers:

```
server/
├── cmd/server/main.go                    # Wiring & bootstrap
└── internal/
    ├── shared/                           # SHARED KERNEL
    │   ├── valueobjects/                 # Money, Weight, IDs
    │   ├── events/                       # DomainEvent interface, BaseEvent
    │   ├── policy/                       # DetectionPolicy
    │   └── errors/                       # Shared domain errors
    │
    ├── catalog/                          # CATALOG BOUNDED CONTEXT
    │   ├── domain/                       # SKU aggregate, repository port
    │   ├── app/                          # CreateSKU use case, queries
    │   ├── infra/                        # Postgres repo, HTTP handlers
    │   └── api/                          # SKUReader interface for cross-context reads
    │
    ├── device/                           # DEVICE BOUNDED CONTEXT
    │   ├── domain/                       # Device aggregate, repository port
    │   ├── app/                          # RegisterDevice use case, queries
    │   ├── infra/                        # Postgres repo, HTTP handlers
    │   └── api/                          # DeviceReader interface for cross-context reads
    │
    ├── transaction/                      # TRANSACTION BOUNDED CONTEXT
    │   ├── domain/                       # Session aggregate, DetectedItem VO
    │   ├── app/                          # StartSession, SubmitDetection, etc.
    │   │   └── ports/                    # Interfaces for cross-context deps
    │   ├── infra/
    │   │   └── adapters/                 # Implements ports using other APIs
    │   └── api/                          # SessionReader interface
    │
    ├── platform/                         # SHARED INFRASTRUCTURE
    │   ├── http/                         # Router (composes all context routes)
    │   ├── postgres/                     # Migrations
    │   └── messaging/                    # Event publisher
    │
    └── pkg/                              # Shared utilities
        └── logger/
```

### Bounded Contexts

| Context | Responsibility | Aggregates |
|---------|---------------|------------|
| **Catalog** | Product/SKU management | SKU |
| **Device** | Vending machine registration | Device |
| **Transaction** | Customer session workflow | Session |

### Cross-Context Communication

Transaction context reads from Catalog and Device contexts via their `api/` packages:

```
Transaction Context
    │
    ├──[DeviceReader port]──> Device Context API (DeviceReader interface)
    │
    └──[CatalogReader port]──> Catalog Context API (SKUReader interface)
```

### Dependency Rule

Dependencies always point inward within each context:
- **domain/** knows nothing about app or infra
- **app/** knows domain but not infra
- **infra/** depends on both but is depended on by nothing inside

Cross-context communication only through `api/` packages (never import another context's domain or infra).

### Key Patterns

| Pattern | Location | Purpose |
|---------|----------|---------|
| Aggregate Root | `<context>/domain/*.go` | Consistency boundary, owns mutations |
| Value Object | `shared/valueobjects/` | Immutable, value-based equality |
| Repository Port | `<context>/domain/repository.go` | Interface defined by domain |
| Repository Adapter | `<context>/infra/postgres_repo.go` | Implements domain interface |
| Use Case Handler | `<context>/app/*.go` | Orchestrates: load → mutate → save → publish |
| Domain Event | `<context>/domain/events.go` | Immutable facts, past-tense names |
| Cross-Context Port | `<context>/app/ports/*.go` | Interface for reading from other contexts |
| Cross-Context Adapter | `<context>/infra/adapters/*.go` | Implements port using other context's API |

### Key API Endpoints

| Method | Path | Context | Description |
|--------|------|---------|-------------|
| POST | `/api/v1/skus` | Catalog | Create SKU (admin) |
| GET | `/api/v1/skus` | Catalog | List all SKUs |
| GET | `/api/v1/skus/:id` | Catalog | Get SKU by ID |
| POST | `/api/v1/device/register` | Device | Register ESP32 device |
| GET | `/api/v1/device/skus` | Device | Get active SKUs (for device sync) |
| POST | `/api/v1/device/detection` | Transaction | Submit detection results |
| POST | `/api/v1/session/start` | Transaction | Start session via QR code |
| GET | `/api/v1/session/:id` | Transaction | Get session details |
| POST | `/api/v1/session/:id/confirm` | Transaction | Confirm purchase |
| POST | `/api/v1/session/:id/cancel` | Transaction | Cancel session |

### Recognition Flow

1. Weight change detected → camera capture
2. On-device object detection (TFLite)
3. Weight cross-validation (sum of detected items vs measured)
4. If confidence < 80% or weight mismatch → upload to cloud
5. Return best result; customer can request refund if wrong

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| DATABASE_URL | postgres://... | PostgreSQL connection |

## Hardware Pins (ESP32-S3-CAM)

- HX711 DT: GPIO1
- HX711 SCK: GPIO2
- Camera pins defined in `firmware/src/main.cpp` (board-specific)
