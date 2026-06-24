# LogMind Documentation

LogMind turns log streams into active incidents with a lightweight Go backend, SQLite persistence, and a small in-memory detection engine.

## Start Here

- [Getting Started](usage.md)
- [Examples](examples.md)
- [Contributing](contributing.md)
- [Decision Log](decisions/index.md)

## What The Service Does

- Accepts logs through `POST /logs`
- Stores logs in SQLite
- Rebuilds recent incident state on startup from persisted error logs
- Exposes `/incidents`, `/healthz`, `/readyz`, and `/metrics`

## When To Read What

- Use `usage.md` to get the service running and send the first logs
- Use `examples.md` if you want runnable sample clients
- Use `contributing.md` if you want to change the codebase safely
