# Incident Recovery

## Decision

Incident state is rebuilt from recent SQLite logs on startup instead of living only in memory.

## Why

The incident engine is intentionally lightweight, but the previous design lost all active incidents on restart because the state only existed in RAM. The logs were already persisted to SQLite, so replaying recent error logs gives us recovery without adding a second datastore or a separate incident table.

## What Changed

- Query recent error logs from SQLite during bootstrap.
- Replay those errors into the in-memory engine before the API starts serving traffic.
- Keep the hot path simple: writes still go to the queue and SQLite, while recovery happens only at startup.

## Tradeoff

Recovery is bounded by the sliding window, so this does not create durable incident history. It does make the live `/incidents` view consistent across restarts, which is the part users actually depend on.
