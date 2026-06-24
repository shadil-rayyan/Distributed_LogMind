# LogMind

LogMind is a Go log-ingestion service that stores logs in SQLite and turns recent error bursts into active incidents.

## Read First

- [Documentation home](Doc/index.md)
- [Getting started](Doc/usage.md)
- [Runnable examples](Doc/examples.md)
- [Contributing guide](Doc/contributing.md)
- [Decision log](Doc/decisions/index.md)

## Quick Start

```bash
docker compose up -d
curl -X POST http://localhost:8080/logs \
  -H "Content-Type: application/json" \
  -d '{"service":"payment-api","level":"error","message":"Database connection failed"}'
curl http://localhost:8080/incidents
```

## Local Development

```bash
go run ./cmd/logmind
go test ./...
go test -race ./...
```

## Documentation Site

The MkDocs source lives in `Doc/`, with the site configuration in `mkdocs.yml`.

Preview it locally with:

```bash
mkdocs serve
```
