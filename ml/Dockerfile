# ML Detection Server
# Multi-stage build for smaller production image

FROM python:3.11-slim AS base

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgl1-mesa-glx \
    libglib2.0-0 \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Poetry
ENV POETRY_VERSION=1.7.1 \
    POETRY_HOME=/opt/poetry \
    POETRY_NO_INTERACTION=1 \
    POETRY_VIRTUALENVS_IN_PROJECT=1 \
    POETRY_VIRTUALENVS_CREATE=1
RUN curl -sSL https://install.python-poetry.org | python3 -
ENV PATH="${POETRY_HOME}/bin:$PATH"

# Create non-root user
RUN useradd --create-home --shell /bin/bash appuser

# Builder stage
FROM base AS builder

WORKDIR /app

# Copy dependency files
COPY pyproject.toml poetry.lock* ./

# Install dependencies
RUN poetry install --only main --no-root --no-ansi

# Production stage
FROM base AS production

WORKDIR /app

# Copy virtual environment from builder
COPY --from=builder /app/.venv /app/.venv

# Copy application code
COPY proto/ /app/proto/
COPY server/ /app/server/

# Copy generated protobuf code (if pre-generated)
# COPY server/generated/ /app/server/generated/

# Set environment
ENV PATH="/app/.venv/bin:$PATH" \
    PYTHONPATH=/app/server \
    PYTHONUNBUFFERED=1 \
    VIRTUAL_ENV=/app/.venv

# Create models directory
RUN mkdir -p /app/models && chown -R appuser:appuser /app

USER appuser

# gRPC port
EXPOSE 50051

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import grpc; from app.generated import detection_pb2_grpc, detection_pb2; \
        channel = grpc.insecure_channel('localhost:50051'); \
        stub = detection_pb2_grpc.DetectionServiceStub(channel); \
        stub.HealthCheck(detection_pb2.Empty(), timeout=5)" || exit 1

CMD ["python", "-m", "app.main"]
