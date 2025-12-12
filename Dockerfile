# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags '-extldflags "-static" -s -w' \
    -o miniflux-auto-read .

# Final stage - using scratch for minimal image size
FROM scratch

# Copy CA certificates for HTTPS requests to Miniflux API
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /build/miniflux-auto-read /miniflux-auto-read

# Expose port
EXPOSE 8080

# Set default environment
ENV PORT=8080

# Run as non-root would require adding user in scratch, so we skip it
# The application doesn't require root privileges

# Run the application
ENTRYPOINT ["/miniflux-auto-read"]
