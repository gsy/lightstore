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

### Backend DDD Hexagonal Architecture

The backend follows Domain-Driven Design with Hexagonal Architecture (Ports & Adapters):

```
server/
├── cmd/server/main.go                    # Wiring & bootstrap
└── internal/
    ├── domain/                           # INNER CORE - Business logic
    │   ├── beverage/                     # Beverage aggregate
    │   │   ├── aggregate.go              # Aggregate root
    │   │   ├── events.go                 # Domain events
    │   │   ├── errors.go                 # Domain errors
    │   │   └── repository.go             # Repository PORT (interface)
    │   ├── device/                       # Device aggregate
    │   ├── session/                      # Session aggregate
    │   └── shared/                       # Shared value objects
    │       └── value_objects.go          # Money, Weight, IDs
    │
    ├── application/                      # MIDDLE RING - Use cases
    │   ├── ports/                        # Output port interfaces
    │   │   ├── event_publisher.go
    │   │   └── ml_service.go
    │   ├── createbeverage/               # Use case: Create beverage
    │   │   ├── command.go                # Input DTO
    │   │   └── handler.go                # Use case handler
    │   ├── registerdevice/               # Use case: Register device
    │   ├── startsession/                 # Use case: Start session
    │   └── submitdetection/              # Use case: Submit detection
    │
    └── infrastructure/                   # OUTER RING - Adapters
        ├── persistence/postgres/         # Repository adapters
        │   ├── beverage_repo.go
        │   ├── device_repo.go
        │   ├── session_repo.go
        │   └── migrations.go
        ├── http/                         # HTTP adapters
        │   ├── server.go                 # Router setup
        │   └── handlers/                 # HTTP handlers
        └── messaging/                    # Event publisher adapters
            └── noop_publisher.go
```

### Dependency Rule

Dependencies always point inward:
- **Domain** knows nothing about application or infrastructure
- **Application** knows domain but not infrastructure
- **Infrastructure** depends on both but is depended on by nothing inside

### Key Patterns

| Pattern | Location | Purpose |
|---------|----------|---------|
| Aggregate Root | `domain/*/aggregate.go` | Consistency boundary, owns mutations |
| Value Object | `domain/shared/` | Immutable, value-based equality (Money, Weight, IDs) |
| Repository Port | `domain/*/repository.go` | Interface defined by domain |
| Repository Adapter | `infrastructure/persistence/` | Implements domain interface |
| Use Case Handler | `application/*/handler.go` | Orchestrates: load → mutate → save → publish |
| Domain Event | `domain/*/events.go` | Immutable facts, past-tense names |

### Key API Endpoints

- `POST /api/v1/device/register` - Register ESP32 device
- `POST /api/v1/device/detection` - Submit detection results
- `POST /api/v1/session/start` - Start session via QR code (machine_id)
- `POST /api/v1/session/:id/confirm` - Confirm purchase
- `POST /api/v1/beverages` - Create beverage (admin)

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
