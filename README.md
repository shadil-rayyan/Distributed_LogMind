# LogMind 🧠

A lightweight, memory-optimized, zero-dependency distributed log monitoring system. Designed for developers who need to know "what's broken right now" without setting up massive enterprise clusters.

## Features

- **Single Binary**: Compiles down to one executable. No Java, no Elasticsearch, no Kafka required.
- **Ultra-low Memory (≤30MB RSS)**: Safely runs on tiny VPS instances or local machines without dragging them down.
- **O(1) Incident Triage**: Tracks rolling error windows natively in RAM before hitting the disk.
- **Crash-Resilient Queues**: Internal backpressure protects against burst-traffic OOM crashes.
- **Observability Built-in**: Exposes Prometheus metrics out of the box (`/metrics`).
- **DevSecOps Ready**: Includes Docker Compose configurations for both development and hardened production setups, along with GitHub Actions CI/CD pipelines.

## Getting Started

### Quick Start (Development)

Run the local development environment using Docker Compose:

```bash
docker compose up -d
```

### Production Deployment

For production, we use a hardened Docker Compose configuration that enforces strict CPU and memory limits:

```bash
docker compose -f docker-compose.prod.yml up -d
```

## API Usage

### 1. Ingest a Log

```bash
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{"service": "payment-api", "level": "error", "message": "Database connection failed"}'
```

### 2. View Active Incidents

```bash
curl http://localhost:8080/incidents
```

### 3. Check System Health & Metrics

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/metrics
```

## Development & Testing

This project is built using a clean `cmd/` and `internal/` architecture. 

To run the unit and integration tests:

```bash
go test ./... -v
```

## Security

We run automated security scans using `gosec`, `govulncheck`, and `trivy` in our CI pipelines. For production deployments, the container runs in read-only mode, drops all capabilities, and restricts resource consumption.

## License

MIT
