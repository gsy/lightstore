.PHONY: help proto server train validate export clean setup test lint docker-build docker-run

# Default target
help:
	@echo "ML Server Commands:"
	@echo "  make setup          - Install dependencies"
	@echo "  make proto          - Generate protobuf code"
	@echo "  make server         - Run ML server"
	@echo "  make test           - Run tests"
	@echo "  make lint           - Run linters"
	@echo ""
	@echo "Training Commands:"
	@echo "  make train          - Train YOLOv8 model"
	@echo "  make validate       - Validate trained model"
	@echo "  make export         - Export model for deployment"
	@echo "  make prepare-data   - Prepare dataset from annotations"
	@echo ""
	@echo "CVAT Commands:"
	@echo "  make cvat-export    - Export annotations from CVAT"
	@echo "  make sync-classes   - Sync class names from catalog"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build   - Build server Docker image"
	@echo "  make docker-run     - Run server container"
	@echo "  make docker-train   - Run training in container"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean          - Clean generated files"

# Setup
setup:
	pip install -r requirements.txt

setup-train:
	pip install -r requirements-train.txt

# Protobuf generation
PROTO_DIR := proto
PROTO_OUT := server/app/generated
GO_PROTO_OUT := ../server/internal/platform/mlclient/generated

proto:
	@mkdir -p $(PROTO_OUT)
	python -m grpc_tools.protoc \
		-I$(PROTO_DIR) \
		--python_out=$(PROTO_OUT) \
		--grpc_python_out=$(PROTO_OUT) \
		$(PROTO_DIR)/detection.proto
	@touch $(PROTO_OUT)/__init__.py
	@echo "Python protobuf code generated in $(PROTO_OUT)"

proto-go:
	@mkdir -p $(GO_PROTO_OUT)
	protoc \
		-I$(PROTO_DIR) \
		--go_out=$(GO_PROTO_OUT) \
		--go-grpc_out=$(GO_PROTO_OUT) \
		$(PROTO_DIR)/detection.proto
	@echo "Go protobuf code generated in $(GO_PROTO_OUT)"

# Server
server: proto
	PYTHONPATH=server python -m app.main

server-debug: proto
	DEBUG=true LOG_LEVEL=DEBUG PYTHONPATH=server python -m app.main

# Training
EPOCHS ?= 100
BATCH ?= 16
MODEL ?= yolov8s.pt
DATA ?= training/configs/beverages.yaml

train:
	python training/scripts/train.py \
		--model $(MODEL) \
		--data $(DATA) \
		--epochs $(EPOCHS) \
		--batch $(BATCH) \
		--augment

train-resume:
	@if [ -z "$(CHECKPOINT)" ]; then \
		echo "Error: CHECKPOINT not set. Usage: make train-resume CHECKPOINT=runs/train/exp/weights/last.pt"; \
		exit 1; \
	fi
	python training/scripts/train.py --resume $(CHECKPOINT)

validate:
	@if [ -z "$(MODEL_PATH)" ]; then \
		echo "Error: MODEL_PATH not set. Usage: make validate MODEL_PATH=runs/train/exp/weights/best.pt"; \
		exit 1; \
	fi
	python training/scripts/validate.py \
		--model $(MODEL_PATH) \
		--data $(DATA) \
		--plots

export:
	@if [ -z "$(MODEL_PATH)" ]; then \
		echo "Error: MODEL_PATH not set. Usage: make export MODEL_PATH=runs/train/exp/weights/best.pt"; \
		exit 1; \
	fi
	python training/scripts/export.py \
		--model $(MODEL_PATH) \
		--output-dir models \
		--format pytorch

# Dataset
prepare-data:
	python scripts/prepare_dataset.py \
		--images-dir data/raw \
		--annotations-dir data/annotations \
		--output-dir data/dataset

# CVAT
CVAT_PROJECT_ID ?= 1

cvat-export:
	@if [ -z "$(CVAT_API_TOKEN)" ]; then \
		echo "Error: CVAT_API_TOKEN not set"; \
		exit 1; \
	fi
	python scripts/cvat_export.py \
		--project-id $(CVAT_PROJECT_ID) \
		--output-dir data/annotations

sync-classes:
	python scripts/sync_classes.py \
		--update-config \
		--update-server

# Testing
test:
	pytest server/tests/ -v --cov=server/app

lint:
	ruff check server/ training/ scripts/
	black --check server/ training/ scripts/
	mypy server/app/

format:
	ruff check --fix server/ training/ scripts/
	black server/ training/ scripts/

# Docker
DOCKER_IMAGE := vending-ml-server
DOCKER_TRAIN_IMAGE := vending-ml-train

docker-build:
	docker build -t $(DOCKER_IMAGE):latest .

docker-build-train:
	docker build -f Dockerfile.train -t $(DOCKER_TRAIN_IMAGE):latest .

docker-run:
	docker run -d \
		--name ml-server \
		-p 50051:50051 \
		-v $(PWD)/models:/app/models \
		$(DOCKER_IMAGE):latest

docker-run-gpu:
	docker run -d \
		--name ml-server \
		--gpus all \
		-p 50051:50051 \
		-v $(PWD)/models:/app/models \
		$(DOCKER_IMAGE):latest

docker-train:
	docker run --rm \
		--gpus all \
		-v $(PWD)/data:/workspace/data \
		-v $(PWD)/models:/workspace/models \
		-v $(PWD)/runs:/workspace/runs \
		-v $(PWD)/training/configs:/workspace/training/configs \
		$(DOCKER_TRAIN_IMAGE):latest \
		--data /workspace/training/configs/beverages.yaml \
		--epochs $(EPOCHS) \
		--batch $(BATCH)

docker-stop:
	docker stop ml-server || true
	docker rm ml-server || true

# Cleanup
clean:
	rm -rf $(PROTO_OUT)/*.py
	rm -rf runs/
	rm -rf __pycache__ **/__pycache__
	rm -rf .pytest_cache .ruff_cache .mypy_cache
	find . -name "*.pyc" -delete
