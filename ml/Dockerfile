# ML Detection Server
# Multi-stage build for smaller production image

FROM python:3.11-slim AS base

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgl1-mesa-glx \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN useradd --create-home --shell /bin/bash appuser

# Builder stage
FROM base AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

# Production stage
FROM base AS production

# Copy installed packages from builder
COPY --from=builder /root/.local /home/appuser/.local

# Copy application code
COPY proto/ /app/proto/
COPY server/ /app/server/

# Copy generated protobuf code (if pre-generated)
# COPY server/generated/ /app/server/generated/

# Set environment
ENV PATH=/home/appuser/.local/bin:$PATH
ENV PYTHONPATH=/app/server
ENV PYTHONUNBUFFERED=1

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
