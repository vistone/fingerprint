# Fingerprint Gateway Docker Image
# Multi-stage build for minimal production image
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
RUN go work sync && go mod download

# Copy remaining source code
COPY . .

# Build the gateway binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o fingerprint-gateway \
    ./cmd/gateway

# Runtime stage — minimal scratch image
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /build/fingerprint-gateway /fingerprint-gateway

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/fingerprint-gateway", "-health"]

USER 65534:65534

ENTRYPOINT ["/fingerprint-gateway"]
