# Contributing

This page is the baseline checklist for anyone changing LogMind.

## Prerequisites

- Go 1.23 or newer
- Docker and Docker Compose
- `mkdocs` if you want to preview the docs site locally

## Run The Service

Local development:

```bash
docker compose up
```

Directly with Go:

```bash
go run ./cmd/logmind
```

Production-style compose:

```bash
docker compose -f docker-compose.prod.yml up -d
```

## Test The Code

Run the full suite:

```bash
go test ./...
```

Run the race detector on the concurrent paths:

```bash
go test -race ./...
```

If you only need the core runtime checks:

```bash
go test ./pkg/ratelimit ./tests/integration ./tests/unit
```

## Format And Verify

Before opening a change:

```bash
gofmt -w <touched-go-files>
go test ./...
```

If you touch docs, preview them:

```bash
mkdocs serve
```

## What To Update When Behavior Changes

- Update the relevant code path
- Update or add a focused test
- Update the usage example in `usage.md` if the user flow changes
- Update the runnable sample in `examples.md` if the client contract changes
- Update the decision log if the change reflects a durable design choice
