# Fingerprint Gateway Docker Image
# Multi-stage build for minimal production image

# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy workspace and module definition files
COPY go.work go.work.sum ./
COPY go.mod go.sum ./

# Copy all submodule go.mod/go.sum files (needed for workspace)
COPY examples examples/
COPY modules modules/

# Download dependencies (now all go.mod files are present)
RUN go mod download

# Copy remaining source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o fingerprint-gateway \
    ./cmd/gateway

# Runtime stage
FROM scratch

# Copy certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary
COPY --from=builder /build/fingerprint-gateway /fingerprint-gateway

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/fingerprint-gateway", "-health"]

# Run as non-root
USER 65534:65534

# Entry point
ENTRYPOINT ["/fingerprint-gateway"]
