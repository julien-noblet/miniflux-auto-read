# Stage 1: Build
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates jq yq && \
    go install github.com/google/go-jsonnet/cmd/jsonnet@v0.21.0

# Create a non-root user
RUN adduser -D -g '' -u 10001 appuser

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Regenerate assets from Jsonnet mixin
RUN mkdir -p assets && \
    jsonnet -S monitoring/mixin/mixin.libsonnet > /tmp/mixin.json && \
    jq -er '.grafanaDashboards["miniflux-auto-read.json"]' /tmp/mixin.json > assets/dashboard.json && \
    jq -er '.prometheusAlerts' /tmp/mixin.json | yq -P > assets/alerts.yaml && \
    rm /tmp/mixin.json

# Build the application with optimizations:
# -pgo=auto: Use default.pgo if available for profile-guided optimization
# -trimpath: Remove local file system paths from the binary
# -ldflags: -s -w removes symbol table and debug information
# GOAMD64=v3: Use modern CPU instructions (AVX/AVX2) for amd64
ARG TARGETOS TARGETARCH
RUN if [ "$TARGETARCH" = "amd64" ]; then export GOAMD64=v3; fi && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -pgo=auto \
    -trimpath \
    -ldflags="-s -w" \
    -o miniflux-auto-read .

# Stage 2: Final minimal image
FROM scratch

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy non-root user info
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy the binary
COPY --from=builder /build/miniflux-auto-read /miniflux-auto-read

# Use non-root user
USER 10001:10001

# Expose port
EXPOSE 8080

# Performance tuning for Go runtime in containers
# GOMEMLIMIT: Reserve 10% for overhead (should be adjusted via env in production)
# GOGC: Balanced approach
ENV GOMEMLIMIT=128MiB \
    GOGC=100 \
    PORT=8080

# Run the application
ENTRYPOINT ["/miniflux-auto-read"]
