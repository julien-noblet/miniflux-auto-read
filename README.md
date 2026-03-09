# Miniflux Auto Read

A lightweight HTTP service that automatically marks Miniflux entries as read and saves them for later review.

## Overview

Miniflux Auto Read is a companion service for [Miniflux](https://miniflux.app) that helps you manage your RSS feeds by automatically processing unread entries. It marks them as read while saving them to your starred/saved items, keeping your feed clean while preserving articles for later.

**Note:** This service only marks entries as read and saves them. For pushing articles to external services (Pocket, Instapaper, etc.), use Miniflux's built-in integrations.

## Features

- 🚀 Simple HTTP API with three endpoints
- ✅ Health check endpoint for monitoring
- 📖 Bulk process unread entries with one API call
- 🔒 Secure configuration via environment variables
- 📊 Metrics for Prometheus (at `/metrics`)
- 🤖 Daemon mode for continuous operation
- 📊 Detailed logging and error reporting
- ⚡ Fast and lightweight (written in Go 1.26 with PGO optimization)
- 🐳 Ultra-minimal Docker image (~5MB) based on `scratch`
- 🔒 Secure by default: runs as non-root user in container

## Installation

### From Source

```bash
git clone https://github.com/julien-noblet/miniflux-auto-read.git
cd miniflux-auto-read
go build -pgo=auto -trimpath -ldflags="-s -w"
```

### Using Docker

The project provides an ultra-optimized image (multi-platform `amd64`/`arm64`) leveraging **Go 1.26** with **PGO (Profile Guided Optimization)** for maximum performance.

```bash
docker run -d \
  --name miniflux-auto-read \
  -e MINIFLUX_API_URL="https://miniflux.example.com" \
  -e MINIFLUX_API_TOKEN="your-token" \
  -p 8080:8080 \
  ghcr.io/julien-noblet/miniflux-auto-read:latest
```

The image is built on `scratch` for a minimal footprint (~5MB) and runs as a non-root user (UID 10001).

### Using Go Install

```bash
go install github.com/julien-noblet/miniflux-auto-read@latest
```

## Configuration

The service is configured entirely through environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `MINIFLUX_API_URL` | Yes | - | Your Miniflux instance URL (e.g., `https://miniflux.example.com`) |
| `MINIFLUX_API_TOKEN` | Yes | - | Your Miniflux API token |
| `PORT` | No | `8080` | HTTP server port |
| `DAEMON` | No | `false` | If set to `true`, the service stays running after processing. If `false`, it processes once and exits. |

### Getting Your API Token

1. Log in to your Miniflux instance
2. Go to **Settings** → **API Keys**
3. Create a new API key
4. Copy the generated token

## Usage

### Running the Service

```bash
export MINIFLUX_API_URL="https://miniflux.example.com"
export MINIFLUX_API_TOKEN="your-api-token-here"
export PORT=8080

./miniflux-auto-read
```

### API Endpoints

#### Health Check

Check if the service is running and can connect to Miniflux:

```bash
curl http://localhost:8080/healthz
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-09T14:30:00Z"
}
```

#### Process Entries

Mark all unread entries as read and save them:

```bash
curl -X POST http://localhost:8080/process
```

**Response:**
```json
{
  "processed": 42,
  "errors": 0,
  "total": 42
}
```

#### Metrics

The service exports Prometheus metrics at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

**Common Metrics:**
- `miniflux_entries_processed_total`: Total number of entries processed.
- `miniflux_entries_processing_duration_seconds`: Time taken to process entries.
- `http_requests_total`: Total number of HTTP requests.
- `miniflux_api_duration_seconds`: Duration of calls to the Miniflux API.

## Automation

### Cron Job

Add to your crontab to run every hour:

```bash
0 * * * * curl -X POST http://localhost:8080/process
```

### Docker

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o miniflux-auto-read

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/miniflux-auto-read .
EXPOSE 8080
CMD ["./miniflux-auto-read"]
```

Run with Docker:

```bash
docker build -t miniflux-auto-read .
docker run -d \
  -p 8080:8080 \
  -e MINIFLUX_API_URL="https://miniflux.example.com" \
  -e MINIFLUX_API_TOKEN="your-token" \
  --name miniflux-auto-read \
  miniflux-auto-read
```

### systemd Service

Create `/etc/systemd/system/miniflux-auto-read.service`:

```ini
[Unit]
Description=Miniflux Auto Read Service
After=network.target

[Service]
Type=simple
User=miniflux
WorkingDirectory=/opt/miniflux-auto-read
Environment="MINIFLUX_API_URL=https://miniflux.example.com"
Environment="MINIFLUX_API_TOKEN=your-token"
Environment="PORT=8080"
ExecStart=/opt/miniflux-auto-read/miniflux-auto-read
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable miniflux-auto-read
sudo systemctl start miniflux-auto-read
```

## Project Structure

```
.
├── main.go        # Application entry point
├── config.go      # Configuration and environment variables
├── server.go      # HTTP server setup and lifecycle
├── handlers.go    # HTTP handlers (healthz, process)
├── go.mod         # Go module definition
└── README.md      # This file
```

## Development

### Prerequisites

- Go 1.24 or higher
- Access to a Miniflux instance

### Building

```bash
go build -o miniflux-auto-read
```

### Testing

```bash
# Start the service
./miniflux-auto-read

# In another terminal
# Test health check
curl http://localhost:8080/healthz

# Test processing (if you have unread entries)
curl -X POST http://localhost:8080/process
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Miniflux](https://miniflux.app) - An excellent minimalist RSS reader
- Built with the [Miniflux Go Client](https://github.com/miniflux/v2)

## Support

If you encounter any issues or have questions:
- Open an issue on [GitHub](https://github.com/julien-noblet/miniflux-auto-read/issues)
- Check the logs for detailed error messages

## Roadmap

- [ ] Docker image on github container registry
- [ ] Configurable scheduling (built-in cron)
- [X] Prometheus metrics endpoint
- [ ] Filter entries by feed or category
- [ ] Web UI for manual triggering
