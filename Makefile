.PHONY: all build run test clean docker-build docker-up docker-down k8s-deploy k8s-delete

# =============================================================================
# Variables
# =============================================================================

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
BINARY_NAME=server

DOCKER_IMAGE=vending-machine/server
DOCKER_TAG=latest
COMPOSE=docker compose

# =============================================================================
# Default Target
# =============================================================================

all: build

# =============================================================================
# Go Backend
# =============================================================================

build:
	cd server && $(GOBUILD) -o ../bin/$(BINARY_NAME) ./cmd/server

run:
	cd server && $(GOCMD) run ./cmd/server

test:
	cd server && $(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)

deps:
	cd server && $(GOCMD) mod download
	cd server && $(GOCMD) mod tidy

# =============================================================================
# BDD Tests
# =============================================================================

test-bdd:
	cd server && $(GOTEST) -v ./test/...

test-bdd-smoke:
	cd server && $(GOTEST) -v ./test/... -godog.tags="@smoke"

test-bdd-catalog:
	cd server && $(GOTEST) -v ./test/... -godog.tags="@catalog"

test-bdd-device:
	cd server && $(GOTEST) -v ./test/... -godog.tags="@device"

test-bdd-transaction:
	cd server && $(GOTEST) -v ./test/... -godog.tags="@transaction"

test-bdd-progress:
	cd server && $(GOTEST) -v ./test/... -godog.format=progress

test-bdd-junit:
	cd server && $(GOTEST) -v ./test/... -godog.format=junit > server/test-report.xml

test-bdd-cucumber:
	cd server && $(GOTEST) -v ./test/... -godog.format=cucumber > server/test-report.json

# =============================================================================
# Docker Compose - Core
# =============================================================================

# Start default stack (postgres + server + ml-server)
up:
	$(COMPOSE) up -d

# Start with GPU-enabled ML server
up-gpu:
	$(COMPOSE) --profile gpu up -d

# Start development mode with hot reload
up-dev:
	$(COMPOSE) --profile dev up -d

# Start with admin tools (pgAdmin, Redis)
up-tools:
	$(COMPOSE) --profile tools up -d

# Start full stack (all services)
up-full:
	$(COMPOSE) --profile full up -d

# Stop all services
down:
	$(COMPOSE) --profile full down

# Stop and remove volumes
down-clean:
	$(COMPOSE) --profile full down -v

# View logs
logs:
	$(COMPOSE) logs -f

logs-server:
	$(COMPOSE) logs -f server

logs-ml:
	$(COMPOSE) logs -f ml-server

# Restart specific services
restart-server:
	$(COMPOSE) restart server

restart-ml:
	$(COMPOSE) restart ml-server

# Service status
ps:
	$(COMPOSE) ps -a

# =============================================================================
# Docker Compose - Aliases (backwards compatibility)
# =============================================================================

docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) ./server

docker-up: up

docker-down: down

docker-logs: logs-server

# =============================================================================
# Database
# =============================================================================

db-connect:
	$(COMPOSE) exec postgres psql -U vending -d vending

db-shell:
	$(COMPOSE) exec postgres sh

db-seed:
	@echo "Seeding sample beverages..."
	@curl -s -X POST http://localhost:8080/api/v1/skus \
		-H "Content-Type: application/json" \
		-d '{"name":"Coca-Cola 330ml","sku":"coke-330","price":2.50,"weight_grams":350,"class_id":0}'
	@curl -s -X POST http://localhost:8080/api/v1/skus \
		-H "Content-Type: application/json" \
		-d '{"name":"Sprite 330ml","sku":"sprite-330","price":2.50,"weight_grams":345,"class_id":1}'
	@curl -s -X POST http://localhost:8080/api/v1/skus \
		-H "Content-Type: application/json" \
		-d '{"name":"Fanta Orange 330ml","sku":"fanta-330","price":2.50,"weight_grams":348,"class_id":2}'
	@curl -s -X POST http://localhost:8080/api/v1/skus \
		-H "Content-Type: application/json" \
		-d '{"name":"Water 500ml","sku":"water-500","price":1.50,"weight_grams":510,"class_id":3}'
	@curl -s -X POST http://localhost:8080/api/v1/skus \
		-H "Content-Type: application/json" \
		-d '{"name":"Red Bull 250ml","sku":"redbull-250","price":3.50,"weight_grams":280,"class_id":4}'
	@echo "\nDone!"

db-reset:
	$(COMPOSE) exec postgres psql -U vending -d vending -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@echo "Database reset. Restart server to run migrations."

# =============================================================================
# API Testing
# =============================================================================

api-health:
	@curl -s http://localhost:8080/health | jq .

api-skus:
	@curl -s http://localhost:8080/api/v1/skus | jq .

api-ml-health:
	@python -c "import grpc; from ml.server.app.generated import detection_pb2, detection_pb2_grpc; \
		ch = grpc.insecure_channel('localhost:50051'); \
		stub = detection_pb2_grpc.DetectionServiceStub(ch); \
		r = stub.HealthCheck(detection_pb2.Empty()); \
		print(f'Healthy: {r.healthy}, Model: {r.model_loaded}, Status: {r.status}')"

# =============================================================================
# Kubernetes
# =============================================================================

k8s-deploy:
	kubectl apply -f deploy/k8s/namespace.yaml
	kubectl apply -f deploy/k8s/postgres.yaml
	kubectl apply -f deploy/k8s/server.yaml

k8s-delete:
	kubectl delete -f deploy/k8s/server.yaml
	kubectl delete -f deploy/k8s/postgres.yaml
	kubectl delete -f deploy/k8s/namespace.yaml

# =============================================================================
# ML Server
# =============================================================================

ml-setup:
	cd ml && pip install -r requirements.txt

ml-setup-train:
	cd ml && pip install -r requirements-train.txt

ml-proto:
	cd ml && make proto

ml-proto-go:
	cd ml && make proto-go

ml-server:
	cd ml && make server

ml-server-debug:
	cd ml && make server-debug

ml-train:
	cd ml && make train EPOCHS=$(EPOCHS) BATCH=$(BATCH)

ml-validate:
	cd ml && make validate MODEL_PATH=$(MODEL_PATH)

ml-export:
	cd ml && make export MODEL_PATH=$(MODEL_PATH)

ml-prepare-data:
	cd ml && make prepare-data

ml-cvat-export:
	cd ml && make cvat-export CVAT_PROJECT_ID=$(CVAT_PROJECT_ID)

ml-sync-classes:
	cd ml && make sync-classes

ml-docker-build:
	cd ml && make docker-build

ml-docker-run:
	cd ml && make docker-run

ml-docker-run-gpu:
	cd ml && make docker-run-gpu

ml-docker-stop:
	cd ml && make docker-stop

ml-test:
	cd ml && make test

ml-lint:
	cd ml && make lint

# Run training in Docker with GPU
ml-train-docker:
	$(COMPOSE) --profile train run --rm ml-train

# =============================================================================
# Development Utilities
# =============================================================================

# Setup everything for first time
setup: deps ml-setup ml-proto
	cp -n .env.example .env 2>/dev/null || true
	@echo "Setup complete! Run 'make up' to start services."

# Generate all protobuf code
proto: ml-proto ml-proto-go
	@echo "Protobuf code generated for Python and Go"

# Clean everything
clean-all: clean down-clean
	cd ml && make clean
	rm -rf bin/

# Show running containers and ports
status:
	@echo "=== Running Containers ==="
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
	@echo ""
	@echo "=== Service URLs ==="
	@echo "  Backend API:     http://localhost:8080"
	@echo "  ML Server gRPC:  localhost:50051"
	@echo "  pgAdmin:         http://localhost:5050 (--profile tools)"
	@echo "  Redis Commander: http://localhost:8081 (--profile tools)"

# =============================================================================
# Help
# =============================================================================

help:
	@echo "Vending Machine Development Commands"
	@echo ""
	@echo "Docker Compose:"
	@echo "  make up              Start default stack (postgres + server + ml-server)"
	@echo "  make up-gpu          Start with GPU-enabled ML server"
	@echo "  make up-dev          Start development mode (hot reload)"
	@echo "  make up-tools        Start with admin tools (pgAdmin, Redis)"
	@echo "  make up-full         Start all services"
	@echo "  make down            Stop all services"
	@echo "  make logs            View all logs"
	@echo "  make ps              Show container status"
	@echo ""
	@echo "Development:"
	@echo "  make setup           First-time setup"
	@echo "  make build           Build Go server binary"
	@echo "  make test            Run Go tests"
	@echo "  make test-bdd        Run BDD tests"
	@echo ""
	@echo "Database:"
	@echo "  make db-connect      Connect to PostgreSQL"
	@echo "  make db-seed         Seed sample data"
	@echo "  make db-reset        Reset database"
	@echo ""
	@echo "ML Server:"
	@echo "  make ml-proto        Generate Python protobuf code"
	@echo "  make ml-server       Run ML server locally"
	@echo "  make ml-train        Train YOLOv8 model"
	@echo "  make ml-train-docker Train with Docker + GPU"
	@echo ""
	@echo "Utilities:"
	@echo "  make status          Show running services"
	@echo "  make clean-all       Clean everything"
