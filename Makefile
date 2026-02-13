.PHONY: all build run test clean docker-build docker-up docker-down k8s-deploy k8s-delete

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
BINARY_NAME=server

# Docker parameters
DOCKER_IMAGE=vending-machine/server
DOCKER_TAG=latest

all: build

# Build the server
build:
	cd server && $(GOBUILD) -o ../bin/$(BINARY_NAME) ./cmd/server

# Run locally (requires postgres running)
run:
	cd server && $(GOCMD) run ./cmd/server

# Run tests
test:
	cd server && $(GOTEST) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)

# Download dependencies
deps:
	cd server && $(GOCMD) mod download
	cd server && $(GOCMD) mod tidy

# Docker commands
docker-build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) ./server

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f server

# Kubernetes commands
k8s-deploy:
	kubectl apply -f deploy/k8s/namespace.yaml
	kubectl apply -f deploy/k8s/postgres.yaml
	kubectl apply -f deploy/k8s/server.yaml

k8s-delete:
	kubectl delete -f deploy/k8s/server.yaml
	kubectl delete -f deploy/k8s/postgres.yaml
	kubectl delete -f deploy/k8s/namespace.yaml

# Database commands
db-connect:
	docker-compose exec postgres psql -U vending -d vending

# Seed sample beverages
db-seed:
	@echo "Seeding sample beverages..."
	@curl -X POST http://localhost:8080/api/v1/beverages \
		-H "Content-Type: application/json" \
		-d '{"name":"Coca-Cola 330ml","sku":"coke-330","price":2.50,"weight_grams":350}'
	@curl -X POST http://localhost:8080/api/v1/beverages \
		-H "Content-Type: application/json" \
		-d '{"name":"Sprite 330ml","sku":"sprite-330","price":2.50,"weight_grams":345}'
	@curl -X POST http://localhost:8080/api/v1/beverages \
		-H "Content-Type: application/json" \
		-d '{"name":"Fanta Orange 330ml","sku":"fanta-330","price":2.50,"weight_grams":348}'
	@curl -X POST http://localhost:8080/api/v1/beverages \
		-H "Content-Type: application/json" \
		-d '{"name":"Water 500ml","sku":"water-500","price":1.50,"weight_grams":510}'
	@curl -X POST http://localhost:8080/api/v1/beverages \
		-H "Content-Type: application/json" \
		-d '{"name":"Red Bull 250ml","sku":"redbull-250","price":3.50,"weight_grams":280}'
	@echo "\nDone!"

# API test
api-health:
	curl http://localhost:8080/health

api-beverages:
	curl http://localhost:8080/api/v1/beverages | jq
