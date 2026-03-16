# syntax=docker/dockerfile:1.7
# Fingerprint Gateway Docker Image
# Multi-stage build: Go builder + NVIDIA CUDA Python runtime for GPU training
# For full deployment with monitoring, see deploy/docker/

# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy workspace and module definition files
COPY go.work go.work.sum ./
COPY go.mod go.sum ./

# Copy all submodule sources (needed for workspace)
COPY examples examples/
COPY modules modules/
COPY cmd cmd/

# Download dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go work sync && go mod download

# Copy remaining source code
COPY . .

# Build the gateway binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o fingerprint-gateway \
    ./cmd/gateway

# Runtime stage — NVIDIA CUDA + Python for GPU training
FROM nvidia/cuda:12.6.3-runtime-ubuntu22.04

# Avoid interactive prompts during package install
ENV DEBIAN_FRONTEND=noninteractive

# Install Python 3 and pip
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3 python3-pip ca-certificates curl tzdata && \
    rm -rf /var/lib/apt/lists/*

# Install PyTorch (CUDA 12.6) and numpy
RUN pip3 install --no-cache-dir torch --index-url https://download.pytorch.org/whl/cu126 && \
    pip3 install --no-cache-dir numpy

# Copy Go binary
COPY --from=builder /build/fingerprint-gateway /fingerprint-gateway

# Copy GPU training script
COPY training/gpu_train.py /app/gpu_train.py

# Create models directory
RUN mkdir -p /models /data /tmp

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["curl", "-f", "http://localhost:8080/health"]

ENTRYPOINT ["/fingerprint-gateway"]
