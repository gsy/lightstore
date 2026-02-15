# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.


## Workflow

Before taking any action or writing any code, you must:
1. Summarize your understanding of the task
2. Outline a step-by-step plan
3. Wait for my approval before proceeding

Only begin implementation after I confirm the plan looks correct.

## Project Overview

Beverage vending machine recognition system using ESP32-S3-CAM for computer vision and HX711 load cell for weight verification. Hybrid ML approach: on-device TFLite inference with cloud fallback.

## Commands

### Docker Compose (Recommended)

```bash
# Profiles
make up                 # Default: postgres + server + ml-server (CPU)
make up-gpu             # Use GPU-enabled ML server
make up-dev             # Development mode with hot reload
make up-tools           # Add pgAdmin, Redis, Redis Commander
make up-full            # All services

# Management
make down               # Stop all services
make down-clean         # Stop and remove volumes
make logs               # View all logs
make logs-server        # View server logs only
make ps                 # Show container status
make status             # Show services and URLs

# First-time setup
make setup              # Install deps, generate protos, create .env
```

### Backend Server (Go)

```bash
# Local development (without Docker)
make deps               # Download Go dependencies
make build              # Build binary to bin/server
make run                # Run server (requires postgres)
make test               # Run tests

# Database
make db-connect         # Connect to postgres via psql
make db-seed            # Seed sample beverages
make db-reset           # Reset database schema

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

### ML Server (Python)

```bash
# Setup
make ml-setup           # Install dependencies
make ml-proto           # Generate gRPC code (Python)
make ml-proto-go        # Generate gRPC code (Go client)

# Run server
make ml-server          # Run ML server (port 50051)
make ml-server-debug    # Run with debug logging

# Training pipeline
make ml-prepare-data    # Prepare dataset from CVAT annotations
make ml-train EPOCHS=100 BATCH=16  # Train YOLOv8
make ml-validate MODEL_PATH=runs/train/exp/weights/best.pt
make ml-export MODEL_PATH=runs/train/exp/weights/best.pt

# CVAT integration
export CVAT_API_TOKEN=<your-token>
make ml-cvat-export CVAT_PROJECT_ID=1  # Export annotations from CVAT
make ml-sync-classes    # Sync class names from catalog DB

# Docker
make ml-docker-build    # Build ML server image
make ml-docker-run      # Run ML server container
make ml-docker-run-gpu  # Run with GPU support
```

## Architecture

### Four-Tier System

1. **ESP32-S3-CAM Device** (`firmware/`) - Captures images, measures weight, runs on-device ML inference
2. **Backend Server** (`server/`) - Go/Gin REST API, PostgreSQL, DDD hexagonal architecture
3. **ML Server** (`ml/`) - Python gRPC server for cloud inference fallback (YOLOv8)
4. **Mobile App** (external) - Customer-facing app for QR scanning, payment, refunds

### ML Server Architecture

The ML server provides cloud-based inference when ESP32 confidence is low:

```
ml/
├── proto/                        # gRPC service definition
│   └── detection.proto
├── server/                       # Python gRPC server
│   └── app/
│       ├── main.py              # Server entry point
│       ├── servicer.py          # gRPC service implementation
│       ├── detector.py          # YOLOv8 inference with hot reload
│       ├── config.py            # Configuration
│       └── generated/           # Proto-generated code
├── training/                     # Training pipeline
│   ├── scripts/
│   │   ├── train.py             # YOLOv8 training
│   │   ├── validate.py          # Model validation
│   │   └── export.py            # Export for deployment
│   └── configs/
│       └── beverages.yaml       # Dataset config (20 classes)
├── scripts/                      # Utilities
│   ├── prepare_dataset.py       # CVAT → YOLO format
│   ├── cvat_export.py           # Export from CVAT
│   └── sync_classes.py          # Sync classes from catalog
├── data/                         # Dataset (gitignored)
└── models/                       # Trained models
```

### ML gRPC API

| Method | Description |
|--------|-------------|
| `Detect` | Run inference on image, return detections with SKU IDs |
| `HealthCheck` | Server health and model status |
| `GetModelInfo` | Model version, classes, metrics |
| `SyncClasses` | Update class→SKU mapping from catalog |

### ML Server Features

- **Hot Reload**: Automatically loads new model when file changes
- **Class Sync**: Pulls SKU names from catalog DB for detection labels
- **GPU Support**: Docker with NVIDIA runtime for training/inference

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
4. If confidence < 80% or weight mismatch → upload to ML server (gRPC)
5. ML server runs YOLOv8 inference, returns detections with SKU IDs
6. Return best result; customer can request refund if wrong

```
ESP32-S3-CAM                    Backend Server                 ML Server
     │                               │                              │
     │──── weight change ───────────>│                              │
     │<─── start session ────────────│                              │
     │                               │                              │
     │──── capture image ──────>     │                              │
     │──── TFLite inference ────>    │                              │
     │                               │                              │
     │ [if confidence < 80%]         │                              │
     │──── upload image ────────────>│──── Detect (gRPC) ─────────>│
     │                               │<─── detections + SKU IDs ───│
     │                               │                              │
     │<─── final result ─────────────│                              │
```

## Environment Variables

### Backend Server (Go)

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| DATABASE_URL | postgres://... | PostgreSQL connection |
| ML_SERVER_ADDRESS | localhost:50051 | ML server gRPC address |

### ML Server (Python)

| Variable | Default | Description |
|----------|---------|-------------|
| GRPC_HOST | 0.0.0.0 | gRPC listen host |
| GRPC_PORT | 50051 | gRPC listen port |
| MODEL_DIR | models | Directory containing model files |
| MODEL_NAME | best.pt | Model filename |
| MODEL_WATCH_INTERVAL | 5.0 | Seconds between model file checks |
| DEFAULT_CONFIDENCE | 0.5 | Default confidence threshold |
| LOG_LEVEL | INFO | Logging level |

## Hardware Pins (ESP32-S3-CAM)

- HX711 DT: GPIO1
- HX711 SCK: GPIO2
- Camera pins defined in `firmware/src/main.cpp` (board-specific)
